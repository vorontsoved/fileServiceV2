package fileSystem

import (
	"context"
	fsv1 "github.com/vorontsoved/protosFileService/gen/go/fileservices"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	status2 "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
	"time"
)

type BrowseElements struct {
	Id          int
	Filename    string
	Created_at  time.Time
	Modified_at time.Time
}

type FileService interface {
	Upload(
		ctx context.Context,
		file []byte,
		fileName string,
	) (status bool, err error)
	Browse(
		ctx context.Context,
	) (files []BrowseElements, err error)
	Export(
		ctx context.Context,
		fileId int64,
	) (file []byte, err error)
}

type serverAPI struct {
	log *slog.Logger
	fsv1.UnimplementedFileServiceServer
	fileService        FileService
	maxBrowseSem       *semaphore.Weighted
	maxUploadExportSem *semaphore.Weighted
}

var (
	browseLimit       = 10
	uploadExportLimit = 100
	acquireTimeout    = 30 * time.Second
)

var dict = map[bool]string{
	true:  "Успешная загрузка",
	false: "Ошибка при загрузке файла",
}

func RegisterServerAPI(gRPC *grpc.Server, fileService FileService) {
	srv := &serverAPI{
		fileService:        fileService,
		maxBrowseSem:       semaphore.NewWeighted(int64(browseLimit)),
		maxUploadExportSem: semaphore.NewWeighted(int64(uploadExportLimit)),
	}
	fsv1.RegisterFileServiceServer(gRPC, srv)
}

func (s *serverAPI) Upload(ctx context.Context, req *fsv1.FileUploadRequest) (*fsv1.FileUploadResponse, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx, acquireTimeout)
	defer cancel()

	if err := s.maxUploadExportSem.Acquire(timeoutCtx, 1); err != nil {
		return nil, status2.Error(codes.ResourceExhausted, "Не удалось получить ресурс: превышено время ожидания")
	}
	defer s.maxUploadExportSem.Release(1)

	file := req.GetFileData()
	fileName := req.GetFileName()
	if len(file) == 0 {
		return &fsv1.FileUploadResponse{Success: false, Message: "Invalid file"}, nil
	}

	status, err := s.fileService.Upload(ctx, file, fileName)
	if err != nil {
		return nil, status2.Error(codes.InvalidArgument, "upload Error")
	}
	return &fsv1.FileUploadResponse{Success: status, Message: dict[status]}, nil
}

func (s *serverAPI) Browse(ctx context.Context, req *fsv1.Empty) (*fsv1.FileBrowseResponse, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, acquireTimeout)
	defer cancel()

	if err := s.maxBrowseSem.Acquire(timeoutCtx, 1); err != nil {
		return nil, status2.Error(codes.ResourceExhausted, "Не удалось получить ресурс: превышено время ожидания")
	}
	defer s.maxBrowseSem.Release(1)

	files, err := s.fileService.Browse(ctx)
	if err != nil {
		return &fsv1.FileBrowseResponse{}, err
	}

	var filesProto []*fsv1.FileSummary

	for _, file := range files {
		fileProto := &fsv1.FileSummary{
			FileId:      int64(file.Id),
			Name:        file.Filename,
			DateCreated: timestamppb.New(file.Created_at),
			DateModify:  timestamppb.New(file.Modified_at),
		}
		filesProto = append(filesProto, fileProto)
	}
	return &fsv1.FileBrowseResponse{Files: filesProto}, nil
}

func (s *serverAPI) Export(ctx context.Context, req *fsv1.FileExportRequest) (*fsv1.FileExportResponse, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, acquireTimeout)
	defer cancel()

	if err := s.maxUploadExportSem.Acquire(timeoutCtx, 1); err != nil {
		return nil, status2.Error(codes.ResourceExhausted, "Не удалось получить ресурс: превышено время ожидания")
	}
	defer s.maxUploadExportSem.Release(1)

	id := req.GetFileId()
	file, err := s.fileService.Export(ctx, id)
	if err != nil {
		return &fsv1.FileExportResponse{}, err
	}

	return &fsv1.FileExportResponse{FileData: file}, nil
}
