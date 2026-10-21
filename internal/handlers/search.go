package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Anurag-spec1/novapromptgobackend/internal/httputil"
	"github.com/Anurag-spec1/novapromptgobackend/internal/repository"
)

type SearchHandler struct {
	Repo *repository.ImageRepo
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	tagsParam := r.URL.Query().Get("tags")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var tags []string
	if tagsParam != "" {
		for _, t := range strings.Split(tagsParam, ",") {
			if t = strings.TrimSpace(t); t != "" {
				tags = append(tags, t)
			}
		}
	}

	if q == "" && len(tags) == 0 {
		httputil.JSON(w, http.StatusOK, map[string]any{"items": []any{}, "total": 0})
		return
	}

	images, total, err := h.Repo.Search(r.Context(), tags, q, limit, offset)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", fmt.Sprint(err))
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]any{
		"items":  images,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
