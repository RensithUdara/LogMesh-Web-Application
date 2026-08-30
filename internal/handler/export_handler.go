package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"logmesh/internal/service"
)

type ExportHandler struct {
	exporter *service.ExportService
}

func NewExportHandler(exporter *service.ExportService) *ExportHandler {
	return &ExportHandler{exporter: exporter}
}

func (h *ExportHandler) LogsCSV(c *gin.Context) {
	query, err := parseSearchQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if query.Limit <= 0 {
		query.Limit = 500
	}

	csvBytes, err := h.exporter.CSV(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export logs"})
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="logmesh-logs.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvBytes)
}
