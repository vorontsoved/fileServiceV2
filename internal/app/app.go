package app

import (
	"context"
	fileSystem "fileservice/internal/transport/grpc/fileservice"
	"fmt"
	"github.com/rs/zerolog"
	"os"
)

type FileService struct {
	log     zerolog.Logger
	fileSvr FileSaver
}

type FileSaver interface {
	FileSaver(ctx context.Context, fileName string) (id int, err error)
	GetNameFiles(ctx context.Context) (files []fileSystem.BrowseElements, err error)
	GetFile(ctx context.Context, fileId int64) (file []byte, err error)
}

func New(log zerolog.Logger, fileSaver FileSaver) (*FileService, error) {
	return &FileService{
		log:     log,
		fileSvr: fileSaver,
	}, nil
}

func (f *FileService) Upload(ctx context.Context, file []byte, filename string) (status bool, err error) {
	currentDir, err := os.Getwd()
	if err != nil {
		f.log.Error().Msg("Ошибка в получении текущей директории")
		return false, err
	}
	filePath := fmt.Sprintf("%s/internal/service/storage/saveFiles/%s", currentDir, filename)

	if _, err := os.Stat(filePath); err == nil {
		f.log.Error().Msg("Файл с таким именем уже существует")
		return false, err
	}

	fileHandle, err := os.Create(filePath)
	if err != nil {
		f.log.Error().Msgf("Ошибка при создании файла: %v", err)
		return false, err
	}
	defer fileHandle.Close()

	_, err = fileHandle.Write(file)
	if err != nil {
		f.log.Error().Msgf("Ошибка при записи файла: %v", err)
		return false, err
	}
	if _, err := f.fileSvr.FileSaver(ctx, filename); err != nil {
		f.log.Error().Msgf("Ошибка в FileSaver: %v", err)
		return false, err
	}
	return true, nil
}

func (f *FileService) Browse(ctx context.Context) (files []fileSystem.BrowseElements, err error) {
	resFiles, err := f.fileSvr.GetNameFiles(ctx)
	if err != nil {
		return []fileSystem.BrowseElements{}, err
	}
	return resFiles, nil
}

func (f *FileService) Export(ctx context.Context, fileId int64) (file []byte, err error) {
	readFile, err := f.fileSvr.GetFile(ctx, fileId)
	if err != nil {
		return readFile, err
	}
	return readFile, nil
}
