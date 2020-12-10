package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/morgine/pkg/upload"
	"github.com/morgine/pkg/upload/model"
	"gorm.io/gorm"
)

func NewMultiFileHandlers(
	db *gorm.DB,
	fileDir string,
	getAuthUser func(ctx *gin.Context) (userID int, ok bool),
) (*upload.MultiFileHandlers, error) {
	storage, err := model.NewFileStorage(
		fileDir,
	)
	if err != nil {
		return nil, err
	}
	fileDB, err := model.NewMultiFileDB(db, storage)
	if err != nil {
		return nil, err
	}
	return upload.NewMultiFileHandlers(fileDB, &commonHandlers{getAuthUser: getAuthUser}), nil
}

type commonHandlers struct {
	getAuthUser func(ctx *gin.Context) (userID int, ok bool)
}

func (c *commonHandlers) GetAuthUser(ctx *gin.Context) (userID int, ok bool) {
	return c.getAuthUser(ctx)
}

func (c *commonHandlers) HandleError(ctx *gin.Context, err error) {
	SendError(ctx, err)
}

//type multiFileStorage struct {
//	dir         string
//	getServeUrl func(file string) (string, error)
//}
//
//func newMultiFileStorage(dir string, getServeUrl func(file string) (string, error)) (model.Storage, error) {
//	err := os.MkdirAll(dir, os.ModePerm)
//	if err != nil {
//		return nil, err
//	} else {
//		return &multiFileStorage{
//			dir:         dir,
//			getServeUrl: getServeUrl,
//		}, nil
//	}
//}
//
//func (m *multiFileStorage) CreateFile(file string, data []byte) error {
//	return ioutil.WriteFile(filepath.Join(m.dir, file), data, 0766)
//}
//
//func (m *multiFileStorage) DeleteFile(file string) error {
//	err := os.Remove(filepath.Join(m.dir, file))
//	if err != nil && !os.IsNotExist(err) {
//		return err
//	}
//	return nil
//}
//
//func (m *multiFileStorage) GetFile(file string) (data []byte, err error) {
//	data, err = ioutil.ReadFile(filepath.Join(m.dir, file))
//	if err != nil && !os.IsNotExist(err) {
//		return nil, err
//	}
//	return data, nil
//}
//
//func (m *multiFileStorage) GetServeUrl(file string) (url string, err error) {
//	return m.getServeUrl(file)
//}
