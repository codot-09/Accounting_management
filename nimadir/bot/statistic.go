package bot

import (
	"fmt"
	"os"
	"time"
	"github.com/jung-kurt/gofpdf"
	"github.com/wcharczuk/go-chart/v2"
)

type Stats struct {
	TotalCargo   int
	TotalExpense int
	Profit       int
	WeeklyData   []DailyReport
}

type DailyReport struct {
	Date       string
	CargoSum   int
	ExpenseSum int
}

func getProfitStatus(profit int) string {
	if profit > 0 {
		return fmt.Sprintf("🟢 Foyda: %d so'm", profit)
	} else if profit == 0 {
		return fmt.Sprintf("🟡 Balans: %d so'm", profit)
	}
	return fmt.Sprintf("🔴 Zarar: %d so'm", profit)
}

func getStatistics() Stats {
	cargos := loadCargos()
	expenses := loadExpenses()

	var totalCargo, totalExpense int
	for _, c := range cargos {
		totalCargo += c.Amount
	}
	for _, e := range expenses {
		totalExpense += e.Amount
	}

	weekly := getWeeklyReport(cargos, expenses)
	return Stats{
		TotalCargo:   totalCargo,
		TotalExpense: totalExpense,
		Profit:       totalCargo - totalExpense,
		WeeklyData:   weekly,
	}
}

func getWeeklyReport(cargos []Cargo, expenses []Expense) []DailyReport {
	weekData := []DailyReport{}
	now := time.Now()

	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format("02-01-2006")
		cargoSum := 0
		expenseSum := 0

		for _, c := range cargos {
			if c.Date == day {
				cargoSum += c.Amount
			}
		}
		for _, e := range expenses {
			if e.Date == day {
				expenseSum += e.Amount
			}
		}

		weekData = append(weekData, DailyReport{
			Date:       day,
			CargoSum:   cargoSum,
			ExpenseSum: expenseSum,
		})
	}

	return weekData
}

func generatePDFReport(stats Stats) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 20)
	pdf.SetFillColor(230, 230, 250)
	pdf.CellFormat(190, 15, "Umumiy Hisobot", "1", 1, "C", true, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 10, "Umumiy Ma'lumotlar", "1", 1, "C", true, 0, "")
	pdf.SetFont("Arial", "", 12)

	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(63, 8, "Kirim", "1", 0, "C", true, 0, "")
	pdf.CellFormat(63, 8, "Chiqim", "1", 0, "C", true, 0, "")
	pdf.CellFormat(64, 8, "Foyda / Zarar", "1", 1, "C", true, 0, "")

	pdf.CellFormat(63, 8, fmt.Sprintf("%d so'm", stats.TotalCargo), "1", 0, "C", false, 0, "")
	pdf.CellFormat(63, 8, fmt.Sprintf("%d so'm", stats.TotalExpense), "1", 0, "C", false, 0, "")
	pdf.CellFormat(64, 8, getProfitStatus(stats.Profit), "1", 1, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 10, "Oxirgi 7 kunlik Hisobot", "1", 1, "C", true, 0, "")
	pdf.SetFont("Arial", "B", 12)

	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(63, 8, "Sana", "1", 0, "C", true, 0, "")
	pdf.CellFormat(63, 8, "Kirim", "1", 0, "C", true, 0, "")
	pdf.CellFormat(64, 8, "Chiqim", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	for _, d := range stats.WeeklyData {
		pdf.CellFormat(63, 8, d.Date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(63, 8, fmt.Sprintf("%d so'm", d.CargoSum), "1", 0, "C", false, 0, "")
		pdf.CellFormat(64, 8, fmt.Sprintf("%d so'm", d.ExpenseSum), "1", 1, "C", false, 0, "")
	}

	fileName := fmt.Sprintf("data/pdfs/report_%s.pdf", time.Now().Format("20060102_150405"))
	err := pdf.OutputFileAndClose(fileName)
	return fileName, err
}

func generateChart(stats Stats) (string, error) {
	var xValues []float64
	var cargoValues, expenseValues []float64
	labels := []string{}

	for i, d := range stats.WeeklyData {
		xValues = append(xValues, float64(i))
		cargoValues = append(cargoValues, float64(d.CargoSum))
		expenseValues = append(expenseValues, float64(d.ExpenseSum))
		labels = append(labels, d.Date)
	}

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name: "Kunlar",
			ValueFormatter: func(v interface{}) string {
				i := int(v.(float64))
				if i >= 0 && i < len(labels) {
					return labels[i]
				}
				return ""
			},
		},
		YAxis: chart.YAxis{Name: "So'm"},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name:    "Kirim",
				XValues: xValues,
				YValues: cargoValues,
			},
			chart.ContinuousSeries{
				Name:    "Chiqim",
				XValues: xValues,
				YValues: expenseValues,
			},
		},
	}

	fileName := fmt.Sprintf("data/charts/chart_%s.png", time.Now().Format("20060102_150405"))
	f, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	err = graph.Render(chart.PNG, f)
	return fileName, err
}
