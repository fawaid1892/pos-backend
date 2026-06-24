package handler

import (
	"net/http"
	"time"

	"pos-multi-branch/backend/internal/model"
	"pos-multi-branch/backend/internal/repository"

	"github.com/google/uuid"
)

type ReportHandler struct{}

func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

// GET /api/v1/branches/{id}/reports/sales?start=&end=
func (h *ReportHandler) Sales(w http.ResponseWriter, r *http.Request) {
	branchID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch id"})
		return
	}

	start, end, err := parseTimeRange(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	rows, totalSales, totalDiscount, totalNet, totalTx, err := repository.GetSalesReport(r.Context(), branchID, start, end)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []model.SalesReportRow{}
	}

	resp := model.SalesReportResponse{}
	resp.Period.Start = start.Format(time.RFC3339)
	resp.Period.End = end.Format(time.RFC3339)
	resp.Rows = rows
	resp.TotalSales = totalSales
	resp.TotalDiscount = totalDiscount
	resp.TotalNet = totalNet
	resp.TotalTransactions = totalTx

	writeJSON(w, http.StatusOK, resp)
}

// GET /api/v1/branches/{id}/reports/stock
func (h *ReportHandler) Stock(w http.ResponseWriter, r *http.Request) {
	branchID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch id"})
		return
	}

	rows, err := repository.GetStockReport(r.Context(), branchID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []model.StockReportRow{}
	}

	resp := model.StockReportResponse{
		Rows:       rows,
		TotalItems: len(rows),
	}

	writeJSON(w, http.StatusOK, resp)
}

// GET /api/v1/branches/{id}/reports/profit-loss?start=&end=
func (h *ReportHandler) ProfitLoss(w http.ResponseWriter, r *http.Request) {
	branchID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch id"})
		return
	}

	start, end, err := parseTimeRange(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	rows, summary, err := repository.GetProfitLossReport(r.Context(), branchID, start, end)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []model.ProfitLossRow{}
	}

	resp := model.ProfitLossReportResponse{}
	resp.Period.Start = start.Format(time.RFC3339)
	resp.Period.End = end.Format(time.RFC3339)
	resp.Rows = rows
	resp.Summary = summary

	writeJSON(w, http.StatusOK, resp)
}

// ─── Helper ───

func parseTimeRange(r *http.Request) (time.Time, time.Time, error) {
	q := r.URL.Query()
	startStr := q.Get("start")
	endStr := q.Get("end")

	now := time.Now()

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			start, err = time.Parse("2006-01-02", startStr)
			if err != nil {
				return start, end, err
			}
		}
	} else {
		start = now.Truncate(24 * time.Hour).AddDate(0, 0, -7) // default: last 7 days
	}

	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			end, err = time.Parse("2006-01-02", endStr)
			if err != nil {
				return start, end, err
			}
		}
	} else {
		end = now
	}

	return start, end, nil
}
