package git

import (
	"io"
	"path/filepath"
	"strings"

	cerrors "github.com/TryCadence/Cadence/internal/errors"
	"github.com/TryCadence/Cadence/internal/logging"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type RepositoryOptions struct {
	ExcludeFiles []string
}

type Repository interface {
	GetCommits(opts *CommitOptions) ([]*Commit, error)
	Close() error
}

type CommitPairProvider interface {
	GetCommitPairs(commits []*Commit) ([]*CommitPair, error)
}

type DiffProvider interface {
	GetCommitDiff(fromHash, toHash string) (string, error)
}

type gitRepository struct {
	repo         *git.Repository
	path         string
	excludeFiles []string
	logger       *logging.Logger
}

func OpenRepository(path string, opts *RepositoryOptions) (Repository, error) {
	if path == "" {
		return nil, cerrors.ValidationError("repository path cannot be empty")
	}

	if opts == nil {
		opts = &RepositoryOptions{}
	}

	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, cerrors.GitError("failed to open repository").WithDetails(path).Wrap(err)
	}

	return &gitRepository{
		repo:         r,
		path:         path,
		excludeFiles: opts.ExcludeFiles,
		logger:       logging.Default(),
	}, nil
}

func (r *gitRepository) GetCommits(opts *CommitOptions) ([]*Commit, error) {
	if opts == nil {
		opts = &CommitOptions{}
	}

	var ref *plumbing.Reference
	var err error

	if opts.Branch != "" {
		ref, err = r.repo.Reference(plumbing.ReferenceName("refs/heads/"+opts.Branch), true)
		if err != nil {
			// If the specified branch doesn't exist, fall back to HEAD
			r.logger.Warn("branch not found, using default branch", "branch", opts.Branch)
			ref, err = r.repo.Head()
			if err != nil {
				return nil, cerrors.GitError("failed to get HEAD").Wrap(err)
			}
		}
	} else {
		ref, err = r.repo.Head()
		if err != nil {
			return nil, cerrors.GitError("failed to get HEAD").Wrap(err)
		}
	}

	commitIter, err := r.repo.Log(&git.LogOptions{
		From: ref.Hash(),
	})
	if err != nil {
		return nil, cerrors.GitError("failed to create commit iterator").Wrap(err)
	}
	defer commitIter.Close()

	commits := make([]*Commit, 0)
	count := 0

	err = commitIter.ForEach(func(c *object.Commit) error {
		if opts.MaxDepth > 0 && count >= opts.MaxDepth {
			return io.EOF
		}

		parents := make([]string, len(c.ParentHashes))
		for i, p := range c.ParentHashes {
			parents[i] = p.String()
		}

		commits = append(commits, &Commit{
			Hash:      c.Hash.String(),
			Author:    c.Author.Name,
			Email:     c.Author.Email,
			Timestamp: c.Author.When,
			Message:   c.Message,
			Parents:   parents,
		})

		count++
		return nil
	})

	if err != nil && err != io.EOF {
		return nil, cerrors.GitError("error iterating commits").Wrap(err)
	}

	return commits, nil
}

func (r *gitRepository) GetCommitPairs(commits []*Commit) ([]*CommitPair, error) {
	if len(commits) < 2 {
		return []*CommitPair{}, nil
	}

	pairs := make([]*CommitPair, 0)
	skippedMerge := 0
	skippedTimeDelta := 0
	skippedDiffErr := 0

	for i := 0; i < len(commits)-1; i++ {
		current := commits[i]
		previous := commits[i+1]

		if len(current.Parents) > 1 {
			skippedMerge++
			continue
		}

		timeDelta := current.Timestamp.Sub(previous.Timestamp)
		if timeDelta <= 0 {
			skippedTimeDelta++
			continue
		}

		stats, err := r.getDiffStats(previous.Hash, current.Hash)
		if err != nil {
			skippedDiffErr++
			r.logger.Warn("skipping commit pair: failed to get diff stats",
				"current_hash", current.Hash,
				"previous_hash", previous.Hash,
				"error", err,
			)
			continue
		}

		// Get the actual diff content for AI analysis
		diffContent, err := r.GetCommitDiff(previous.Hash, current.Hash)
		if err != nil {
			// If we can't get diff content, continue without it but log the issue
			r.logger.Warn("failed to get diff content, proceeding without it",
				"current_hash", current.Hash,
				"previous_hash", previous.Hash,
				"error", err,
			)
			diffContent = ""
		}

		pairs = append(pairs, &CommitPair{
			Previous:    previous,
			Current:     current,
			TimeDelta:   timeDelta,
			Stats:       stats,
			DiffContent: diffContent,
		})
	}

	if skippedDiffErr > 0 || skippedMerge > 0 || skippedTimeDelta > 0 {
		r.logger.Info("commit pair generation complete",
			"total_commits", len(commits),
			"pairs_generated", len(pairs),
			"skipped_merge", skippedMerge,
			"skipped_time_delta", skippedTimeDelta,
			"skipped_diff_error", skippedDiffErr,
		)
	}

	return pairs, nil
}

func (r *gitRepository) shouldExcludeFile(filePath string) bool {
	if len(r.excludeFiles) == 0 {
		return false
	}

	for _, pattern := range r.excludeFiles {
		matched, err := filepath.Match(pattern, filepath.Base(filePath))
		if err == nil && matched {
			return true
		}

		matched, err = filepath.Match(pattern, filePath)
		if err == nil && matched {
			return true
		}
	}

	return false
}

func (r *gitRepository) getDiffStats(fromHash, toHash string) (*DiffStats, error) {
	fromCommit, err := r.repo.CommitObject(plumbing.NewHash(fromHash))
	if err != nil {
		return nil, cerrors.GitError("failed to get from commit").WithDetails(fromHash).Wrap(err)
	}

	toCommit, err := r.repo.CommitObject(plumbing.NewHash(toHash))
	if err != nil {
		return nil, cerrors.GitError("failed to get to commit").WithDetails(toHash).Wrap(err)
	}

	fromTree, err := fromCommit.Tree()
	if err != nil {
		return nil, cerrors.GitError("failed to get from tree").Wrap(err)
	}

	toTree, err := toCommit.Tree()
	if err != nil {
		return nil, cerrors.GitError("failed to get to tree").Wrap(err)
	}

	changes, err := fromTree.Diff(toTree)
	if err != nil {
		return nil, cerrors.GitError("failed to get diff").Wrap(err)
	}

	stats := &DiffStats{}
	filesChanged := make(map[string]bool)
	filesChangedTotal := make(map[string]bool)

	for _, change := range changes {
		patch, err := change.Patch()
		if err != nil {
			continue
		}

		for _, filePatch := range patch.FilePatches() {
			from, to := filePatch.Files()

			var filePath string
			if to != nil {
				filePath = to.Path()
			} else if from != nil {
				filePath = from.Path()
			}

			if from != nil {
				filesChangedTotal[from.Path()] = true
			}
			if to != nil {
				filesChangedTotal[to.Path()] = true
			}

			isExcluded := r.shouldExcludeFile(filePath)

			if !isExcluded {
				if from != nil {
					filesChanged[from.Path()] = true
				}
				if to != nil {
					filesChanged[to.Path()] = true
				}
			}

			chunks := filePatch.Chunks()
			for _, chunk := range chunks {
				lines := strings.Split(chunk.Content(), "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					switch chunk.Type() {
					case diff.Add:
						stats.TotalAdditions++
						if !isExcluded {
							stats.Additions++
						}
					case diff.Delete:
						stats.TotalDeletions++
						if !isExcluded {
							stats.Deletions++
						}
					}
				}
			}
		}
	}

	stats.FilesChanged = len(filesChanged)
	stats.FilesChangedTotal = len(filesChangedTotal)

	return stats, nil
}

func (r *gitRepository) Close() error {
	return nil
}

func (r *gitRepository) GetCommitDiff(fromHash, toHash string) (string, error) {
	fromCommit, err := r.repo.CommitObject(plumbing.NewHash(fromHash))
	if err != nil {
		return "", cerrors.GitError("failed to get from commit").WithDetails(fromHash).Wrap(err)
	}

	toCommit, err := r.repo.CommitObject(plumbing.NewHash(toHash))
	if err != nil {
		return "", cerrors.GitError("failed to get to commit").WithDetails(toHash).Wrap(err)
	}

	fromTree, err := fromCommit.Tree()
	if err != nil {
		return "", cerrors.GitError("failed to get from tree").Wrap(err)
	}

	toTree, err := toCommit.Tree()
	if err != nil {
		return "", cerrors.GitError("failed to get to tree").Wrap(err)
	}

	changes, err := fromTree.Diff(toTree)
	if err != nil {
		return "", cerrors.GitError("failed to get diff").Wrap(err)
	}

	var diffContent strings.Builder
	for _, change := range changes {
		patch, err := change.Patch()
		if err != nil {
			continue
		}

		filePatches := patch.FilePatches()
		if len(filePatches) == 0 {
			continue
		}

		// Check first file patch to determine if we should exclude
		from, to := filePatches[0].Files()
		var filePath string
		if to != nil {
			filePath = to.Path()
		} else if from != nil {
			filePath = from.Path()
		}

		if r.shouldExcludeFile(filePath) {
			continue
		}

		diffContent.WriteString(patch.String())
	}

	return diffContent.String(), nil
}
