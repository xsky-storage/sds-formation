package formation

import (
	"fmt"
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

type fsQuotaTreeReq struct {
	// name of quota tree
	Name string `json:"name" required:"true"`
	// size of quota tree
	Size uint64 `json:"size,omitempty"`
	// soft quota size of quota tree
	SoftQuotaSize uint64 `json:"soft_quota_size,omitempty"`
}

// FSFolderAddQuotaTreesReq defines request for adding quota trees to folder
type FSFolderAddQuotaTreesReq struct {
	Folder struct {
		QuotaTrees []*fsQuotaTreeReq `json:"fs_quota_trees" required:"true"`
	} `json:"fs_folder" required:"true"`
}

// FSFolderQuotaTree resource
type FSFolderQuotaTree struct {
	ResourceBase

	FolderID      *parser.IntegerExpr
	Name          *parser.StringExpr
	Size          *parser.IntegerExpr
	SoftQuotaSize *parser.IntegerExpr
}

// Init inits resource instance
func (quotaTree *FSFolderQuotaTree) Init(stack utils.StackInterface) {
	quotaTree.ResourceBase.Init(stack)
	quotaTree.setDelegate(quotaTree)
}

// GetType return resource type
func (quotaTree *FSFolderQuotaTree) GetType() string {
	return utils.ResourceFSQuotaTree
}

// IsReady check if the formation args are ready
func (quotaTree *FSFolderQuotaTree) IsReady() (ready bool) {
	if !quotaTree.isReady(quotaTree.FolderID) ||
		!quotaTree.isReady(quotaTree.Name) ||
		!quotaTree.isReady(quotaTree.Size) ||
		!quotaTree.isReady(quotaTree.SoftQuotaSize) {
		return false
	}

	return true
}

func (quotaTree *FSFolderQuotaTree) fakeCreate() (bool, error) {
	quotaTree.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (quotaTree *FSFolderQuotaTree) Create() (created bool, err error) {
	if quotaTree.Name == nil {
		err = errors.Errorf("Name is required for resource %s", quotaTree.GetType())
		return
	}
	if quotaTree.FolderID == nil {
		err = errors.Errorf("FolderID is required for resource %s", quotaTree.GetType())
		return
	}
	if config.DryRun {
		return quotaTree.fakeCreate()
	}

	name := quotaTree.getStringValue(quotaTree.Name)
	folderID := quotaTree.getIntegerValue(quotaTree.FolderID)
	params := map[string]string{"fs_folder_id": fmt.Sprintf("%d", folderID)}
	resourceID, err := quotaTree.getResourceByName(name, params)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		quotaTree.repr = resourceID
		return false, nil
	}

	req := new(FSFolderAddQuotaTreesReq)
	tree := new(fsQuotaTreeReq)
	req.Folder.QuotaTrees = append(req.Folder.QuotaTrees, tree)
	if quotaTree.Name != nil {
		tree.Name = quotaTree.getStringValue(quotaTree.Name)
	}
	if quotaTree.Size != nil {
		tree.Size = uint64(quotaTree.getIntegerValue(quotaTree.Size))
	}
	if quotaTree.SoftQuotaSize != nil {
		tree.SoftQuotaSize = uint64(quotaTree.getIntegerValue(quotaTree.SoftQuotaSize))
	}

	pathParam := map[string]string{"fs_folder_id": fmt.Sprintf("%d", folderID)}
	_, err = quotaTree.CallCreateAPI(req, pathParam)
	if err != nil {
		return false, errors.Annotatef(err, "create fs quota tree %s", name)
	}

	resourceID, err = quotaTree.getResourceByName(name, params)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		quotaTree.repr = resourceID
	} else {
		return false, errors.Errorf("create fs quota tree %s for fs folder %d failed", name, folderID)
	}

	return false, nil
}
