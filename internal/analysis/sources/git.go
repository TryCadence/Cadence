package sources

import (
	"context"
	"fmt"
	"os"

	"github.com/TryCadence/Cadence/internal/analysis"
	"github.com/TryCadence/Cadence/internal/analysis/adapters/git"
)

type GitRepositorySource struct {
	Path   string
	Branch string
}

func NewGitRepositorySource(path, branch string) *GitRepositorySource {
	return &GitRepositorySource{
		Path:   path,
		Branch: branch,
	}
}

func (g *GitRepositorySource) Type() string {
	return "git"
}

func (g *GitRepositorySource) Validate(ctx context.Context) error {
	if g.Path == "" {
		return fmt.Errorf("repository path is required")
	}

	info, err := os.Stat(g.Path)
	if err != nil {
		return fmt.Errorf("repository path does not exist: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("repository path is not a directory")
	}

	return nil
}

func (g *GitRepositorySource) Fetch(ctx context.Context) (*analysis.SourceData, error) {
	repo, err := git.OpenRepository(g.Path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	defer repo.Close()

	opts := &git.CommitOptions{}
	if g.Branch != "" {
		opts.Branch = g.Branch
	}

	commits, err := repo.GetCommits(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	provider, ok := repo.(git.CommitPairProvider)
	if !ok {
		return nil, fmt.Errorf("repository does not support CommitPairProvider interface")
	}

	pairs, err := provider.GetCommitPairs(commits)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit pairs: %w", err)
	}

	return &analysis.SourceData{
		ID:         g.Path,
		Type:       "git",
		RawContent: pairs,
		Metadata: map[string]interface{}{
			"branch":       g.Branch,
			"commit_count": len(commits),
			"commit_pairs": pairs,
		},
	}, nil
}
