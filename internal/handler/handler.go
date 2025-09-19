package handler

import (
	"mime/multipart"
	"net/http"
	"upload-util/internal/config"
	"upload-util/internal/service"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	factory  *service.UploadFactory
	uploader service.Uploader
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type UploadResponse struct {
	URL      string `json:"url"`
	Key      string `json:"key"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	Filename string `json:"filename"`
}

type DeleteRequest struct {
	Key string `json:"key" binding:"required"`
}

type GetURLResponse struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

func NewUploadHandler(cfg *config.UploadConfig) (*UploadHandler, error) {
	factory := service.NewUploadFactory(cfg)
	uploader, err := factory.CreateUploader()
	if err != nil {
		return nil, err
	}
	return &UploadHandler{
		factory:  factory,
		uploader: uploader,
	}, nil
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			Response{
				Code:    http.StatusBadRequest,
				Message: "获取文件失败" + err.Error(),
			})
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	result, err := h.uploader.Upload(c.Request.Context(), file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    http.StatusInternalServerError,
			Message: "上传文件失败" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "上传成功",
		Data: UploadResponse{
			URL:      result.URL,
			Key:      result.Key,
			Size:     result.Size,
			MimeType: result.MimeType,
			Filename: header.Filename,
		},
	})
	return
}

func (h *UploadHandler) UploadMultiple(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    http.StatusBadRequest,
			Message: "解析表单失败: " + err.Error(),
		})
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, Response{
			Code:    http.StatusBadRequest,
			Message: "未找到要上传的文件",
		})
		return
	}
	var results []UploadResponse
	var errList []string
	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			errList = append(errList, "打开文件"+header.Filename+"失败"+err.Error())
			continue
		}
		file.Close()
		result, err := h.uploader.Upload(c.Request.Context(), file, header)
		file.Close()
		if err != nil {
			errList = append(errList, "上传文件"+header.Filename+"失败: "+err.Error())
			continue
		}
		results = append(results, UploadResponse{
			URL:      result.URL,
			Key:      result.Key,
			Size:     result.Size,
			MimeType: result.MimeType,
			Filename: header.Filename,
		})
	}
	response := gin.H{
		"success_count": len(results),
		"error_count":   len(errList),
		"results":       results,
	}
	if len(errList) > 0 {
		response["errors"] = errList
	}

	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "批量上传成功",
		Data:    response,
	})
}

func (h *UploadHandler) Delete(c *gin.Context) {
	var req DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}
	if err := h.uploader.Delete(c.Request.Context(), req.Key); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    http.StatusInternalServerError,
			Message: "删除文件失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "删除成功",
	})
}

func (h *UploadHandler) GetURL(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    http.StatusBadRequest,
			Message: "参数 key 不能为空",
		})
		return
	}

	url, err := h.uploader.GetURL(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    http.StatusInternalServerError,
			Message: "获取URL失败: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "获取成功",
		Data: GetURLResponse{
			URL: url,
			Key: key,
		},
	})
}

func (h *UploadHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "OK",
		Data: gin.H{
			"status":  "ok",
			"service": "upload-util",
		},
	})
}
