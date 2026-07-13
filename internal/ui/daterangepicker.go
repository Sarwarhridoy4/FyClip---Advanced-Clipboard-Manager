package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type dateRangePicker struct {
	widget.BaseWidget
	from         time.Time
	to           time.Time
	viewMonth    time.Time
	selectingEnd bool
	onChanged    func(from, to time.Time)
}

type DateRangePicker struct {
	*dateRangePicker
}

func NewDateRangePicker(initialFrom, initialTo time.Time, onChanged func(from, to time.Time)) *DateRangePicker {
	if initialFrom.IsZero() && initialTo.IsZero() {
		initialFrom = time.Now()
		initialTo = time.Now()
	}
	p := &dateRangePicker{
		from:      initialFrom,
		to:        initialTo,
		viewMonth: initialFrom,
		onChanged: onChanged,
	}
	p.ExtendBaseWidget(p)
	return &DateRangePicker{dateRangePicker: p}
}

func (d *dateRangePicker) CreateRenderer() fyne.WidgetRenderer {
	prevBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		d.viewMonth = d.viewMonth.AddDate(0, -1, 0)
		d.Refresh()
	})
	nextBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		d.viewMonth = d.viewMonth.AddDate(0, 1, 0)
		d.Refresh()
	})
	monthLabel := widget.NewLabel(d.viewMonth.Format("January 2006"))
	monthLabel.Alignment = fyne.TextAlignCenter

	dayHeaders := make([]*widget.Label, 7)
	weekdays := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	for i, wd := range weekdays {
		dayHeaders[i] = widget.NewLabel(wd)
		dayHeaders[i].Alignment = fyne.TextAlignCenter
	}

	dayButtons := make([]*widget.Button, 42)
	start := startOfWeek(d.viewMonth)
	for i := 0; i < 42; i++ {
		current := dayAt(start, i)
		text := ""
		if !current.IsZero() && current.Month() == d.viewMonth.Month() && current.Year() == d.viewMonth.Year() {
			text = fmt.Sprintf("%d", current.Day())
		}
		idx := i
		dayButtons[i] = widget.NewButton(text, func() {
			current := dayAt(d.viewMonth, idx)
			if current.IsZero() {
				return
			}
			if !d.selectingEnd || d.from.IsZero() {
				d.from = current
				d.to = current
				d.selectingEnd = true
			} else {
				if current.Before(d.from) {
					d.to = d.from
					d.from = current
				} else {
					d.to = current
				}
				d.selectingEnd = false
			}
			if d.onChanged != nil {
				d.onChanged(d.from, d.to)
			}
			d.Refresh()
		})
	}

	headerRow := container.NewHBox(prevBtn, monthLabel, nextBtn)
	header := container.NewGridWithColumns(3, prevBtn, container.NewMax(monthLabel), nextBtn)
	headerGrid := container.NewGridWithColumns(7, castToCanvasObjects(dayHeaders, nil)...)
	dayGrid := container.NewGridWithColumns(7, castToCanvasObjects(nil, dayButtons)...)
	content := container.NewVBox(header, headerGrid, dayGrid)

	r := &dateRangePickerRenderer{
		picker:      d,
		monthLabel:  monthLabel,
		dayButtons:  dayButtons,
		dayHeaders:  dayHeaders,
		content:     content,
		prevBtn:     prevBtn,
		nextBtn:     nextBtn,
		headerRow:   headerRow,
		header:      header,
		headerGrid:  headerGrid,
		dayGrid:     dayGrid,
		objects:     []fyne.CanvasObject{content},
	}
	return r
}

type dateRangePickerRenderer struct {
	picker      *dateRangePicker
	monthLabel  *widget.Label
	dayButtons  []*widget.Button
	dayHeaders  []*widget.Label
	content     *fyne.Container
	prevBtn     *widget.Button
	nextBtn     *widget.Button
	headerRow   *fyne.Container
	header      *fyne.Container
	headerGrid  *fyne.Container
	dayGrid     *fyne.Container
	objects     []fyne.CanvasObject
}

func (r *dateRangePickerRenderer) Layout(size fyne.Size) {
	r.content.Resize(size)
}

func (r *dateRangePickerRenderer) MinSize() fyne.Size {
	return r.content.MinSize()
}

func (r *dateRangePickerRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *dateRangePickerRenderer) Destroy() {
}

func (r *dateRangePickerRenderer) Refresh() {
	r.monthLabel.SetText(r.picker.viewMonth.Format("January 2006"))
	r.monthLabel.Refresh()
	start := startOfWeek(r.picker.viewMonth)
	for i := 0; i < 42; i++ {
		current := dayAt(start, i)
		btn := r.dayButtons[i]
		if current.IsZero() || current.Month() != r.picker.viewMonth.Month() || current.Year() != r.picker.viewMonth.Year() {
			btn.SetText("")
			btn.Hide()
			continue
		}
		btn.Show()
		btn.SetText(fmt.Sprintf("%d", current.Day()))
		if !r.picker.from.IsZero() && !r.picker.to.IsZero() {
			if (current.Equal(r.picker.from) || current.After(r.picker.from)) && (current.Equal(r.picker.to) || current.Before(r.picker.to)) {
				btn.Importance = widget.HighImportance
			} else {
				btn.Importance = widget.MediumImportance
			}
		} else if !r.picker.from.IsZero() && isSameDay(current, r.picker.from) {
			btn.Importance = widget.HighImportance
		} else {
			btn.Importance = widget.MediumImportance
		}
	}
	r.dayGrid.Refresh()
	r.headerGrid.Refresh()
	r.header.Refresh()
	r.content.Refresh()
}

func dayAt(base time.Time, idx int) time.Time {
	start := startOfWeek(base)
	return start.AddDate(0, 0, idx)
}

func startOfWeek(t time.Time) time.Time {
	year, month, day := t.Date()
	weekday := time.Date(year, month, day, 0, 0, 0, 0, t.Location()).Weekday()
	offset := int(weekday - time.Sunday)
	if offset < 0 {
		offset += 7
	}
	return time.Date(year, month, day-offset, 0, 0, 0, 0, t.Location())
}

func isSameDay(a, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return false
	}
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func castToCanvasObjects(labels []*widget.Label, buttons []*widget.Button) []fyne.CanvasObject {
	result := make([]fyne.CanvasObject, 0, len(labels)+len(buttons))
	for _, l := range labels {
		result = append(result, l)
	}
	for _, b := range buttons {
		result = append(result, b)
	}
	return result
}
