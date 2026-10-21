package handlers

import (
	"io"
	"net/http"
	"time"

	"github.com/Anurag-spec1/novapromptgobackend/internal/cloudinary"
	"github.com/Anurag-spec1/novapromptgobackend/internal/httputil"
)

type UploadHandler struct {
	Cloudinary *cloudinary.Client
}

func (h *UploadHandler) Signature(w http.ResponseWriter, r *http.Request) {
	folder := r.URL.Query().Get("folder")
	sp, err := h.Cloudinary.SignUpload(folder, time.Hour)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "sign_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, sp)
}

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_multipart", err.Error())
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "missing_file", "expected multipart field 'file'")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "read_file", err.Error())
		return
	}

	res, err := h.Cloudinary.Upload(r.Context(), data, header.Filename)
	if err != nil {
		httputil.Error(w, http.StatusBadGateway, "cloudinary_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, res)
}