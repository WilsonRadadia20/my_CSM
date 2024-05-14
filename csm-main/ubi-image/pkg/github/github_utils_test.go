package github

import (
	"errors"
	"testing"

	log "github.com/sirupsen/logrus"
)

type MockGithubAuth struct {
	BranchExist bool
	DeleteError error
}

func (m *MockGithubAuth) mockGitVerifyBranch() {
	if m.BranchExist {
		log.Infoln("The branch exist, deleting the branch")

		if m.DeleteError != nil {
			log.Errorln("Branch deleting error: ", m.DeleteError.Error())
		} else {
			log.Infoln("The branch deleted successfully")
		}
	} else {
		log.Infoln("The branch does not exist")
	}
}

func TestGithubUtils(t *testing.T) {
	t.Run("Github Utils: Testing", func(t *testing.T) {

		mock1 := &MockGithubAuth{BranchExist: true}
		mock1.mockGitVerifyBranch()

		mock2 := &MockGithubAuth{BranchExist: true, DeleteError: errors.New("Deletion failed")}
		mock2.mockGitVerifyBranch()

		mock3 := &MockGithubAuth{BranchExist: false}
		mock3.mockGitVerifyBranch()
	})
}
