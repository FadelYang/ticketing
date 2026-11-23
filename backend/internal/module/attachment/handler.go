package attachment

import (
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v2"

	"monorepo/backend/internal/shared/middleware"
	"monorepo/backend/internal/shared/response"
)

type Handler struct {
	usecase Usecase
}

func NewHandler(usecase Usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) Upload(c *fiber.Ctx) error {
	userID, _ := c.Locals(middleware.ContextUserIDKey).(string)
	userRole, _ := c.Locals(middleware.ContextUserRoleKey).(string)

	form, err := c.MultipartForm()
	if err != nil {
		return response.BadRequest(c, "multipart form data is required")
	}

	targetType := firstFormValue(form, "type")
	targetID := firstFormValue(form, "target_id")

	fileHeaders := form.File["files"]
	if len(fileHeaders) == 0 {
		fileHeaders = form.File["file"]
	}
	if len(fileHeaders) == 0 {
		return response.BadRequest(c, "at least one image file is required")
	}

	uploads := make([]UploadFileInput, 0, len(fileHeaders))
	openedFiles := make([]multipart.File, 0, len(fileHeaders))
	defer func() {
		for _, f := range openedFiles {
			_ = f.Close()
		}
	}()

	for _, fh := range fileHeaders {
		f, err := fh.Open()
		if err != nil {
			return response.BadRequest(c, "failed to read uploaded file")
		}
		openedFiles = append(openedFiles, f)

		uploads = append(uploads, UploadFileInput{
			FileName:    fh.Filename,
			ContentType: fh.Header.Get("Content-Type"),
			SizeBytes:   fh.Size,
			Reader:      f,
		})
	}

	items, err := h.usecase.Upload(c.UserContext(), userID, userRole, targetType, targetID, uploads)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, fiber.StatusCreated, items)
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, _ := c.Locals(middleware.ContextUserIDKey).(string)
	userRole, _ := c.Locals(middleware.ContextUserRoleKey).(string)

	targetType := c.Query("type")
	var targetIDs []string
	if rawIDs := strings.TrimSpace(c.Query("target_ids")); rawIDs != "" {
		targetIDs = strings.Split(rawIDs, ",")
	} else if rawID := strings.TrimSpace(c.Query("target_id")); rawID != "" {
		targetIDs = []string{rawID}
	}

	items, err := h.usecase.ListByTarget(c.UserContext(), userID, userRole, targetType, targetIDs)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, fiber.StatusOK, items)
}

func (h *Handler) GetFile(c *fiber.Ctx) error {
	userID, _ := c.Locals(middleware.ContextUserIDKey).(string)
	userRole, _ := c.Locals(middleware.ContextUserRoleKey).(string)
	id := c.Params("id")

	att, reader, err := h.usecase.GetFile(c.UserContext(), id, userID, userRole)
	if err != nil {
		return response.Error(c, err)
	}

	c.Set("Content-Type", att.ContentType)
	c.Set("Cache-Control", "private, max-age=3600")
	return c.SendStream(reader)
}

func firstFormValue(form *multipart.Form, key string) string {
	if form == nil || len(form.Value[key]) == 0 {
		return ""
	}
	return form.Value[key][0]
}
