package controller_artworkscreens

import (
	"io/fs"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Anatole-Code/art-expert-backend/internal/artworks/model_artworks"
	"github.com/Anatole-Code/art-expert-backend/internal/artworks/view_artworkscreens"
)

type ArtworkTileView struct {
	ArtworkTitle        string
	AttributionNote      string
	CreatedOn            string
	DiagonalCm           int
	PainterName          string
	ArtworkFileBase     string
	ArtworkImageKey     string
	ArtworkVideoKey     string
	ConnoisseurMarkCount int
	ArtworkFeedURL      string
	NextArtworkFeedURL  string
}

type ArtworkCalendarCell struct {
	Label  string
	Href   string
	Chosen bool
}

type ArtworkScreenData struct {
	Title                string
	ArtworkTab          string
	ArtworkFileBase     string
	ArtworkTiles        []ArtworkTileView
	Artwork             ArtworkTileView
	CreatedOn            string
	CalendarOpen         bool
	CalendarView         string
	CalendarPrev         string
	CalendarNext         string
	CalendarMonthHref    string
	CalendarYearHref     string
	CalendarMonthName    string
	CalendarYearLabel    string
	CalendarCells        []ArtworkCalendarCell
	DraftArtworkTitle   string
	DraftAttributionNote string
	DraftCreatedOn       string
	DraftDiagonal        string
	ArtworkMissing       bool
}

func artworkStyleFS() (fs.FS, error) {
	return fs.Sub(view_artworkscreens.Files, "styles")
}

func (screens *ArtworkScreens) showCreatedOnTile(w http.ResponseWriter, r *http.Request) {
	createdOnQuery := r.URL.Query().Get("created_on")
	createdOn, enabled := parseArtworkDate(createdOnQuery)
	published := screens.artworkLookup.PublishedArtworks(createdOn, enabled)
	shownDate := ""
	if enabled {
		shownDate = createdOn.Format("02.01.2006")
	}
	view := r.URL.Query().Get("calendar")
	if view != "months" && view != "years" {
		view = "days"
	}
	month := artworkCalendarMonth(r.URL.Query().Get("month"), createdOn, enabled, screens.openedOn)
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	if year == 0 {
		year = month.Year()
	}
	blockStart := year - year%12
	mediaBase := screens.mediaBase()
	screen := ArtworkScreenData{
		Title:             "Плитка картин",
		ArtworkTab:        "artworks-created-on-tile",
		ArtworkFileBase:   mediaBase,
		ArtworkTiles:      screens.artworkTileViews(published, mediaBase),
		CreatedOn:         shownDate,
		CalendarOpen:      r.URL.Query().Get("calendar") != "",
		CalendarView:      view,
		CalendarMonthName: artworkMonthNames[month.Month()],
		CalendarYearLabel: strconv.Itoa(year),
		CalendarMonthHref: artworkPickerLink("months", month, year, shownDate),
		CalendarYearHref:  artworkPickerLink("years", month, blockStart, shownDate),
	}
	switch view {
	case "months":
		screen.CalendarPrev = artworkPickerLink("months", month, year-1, shownDate)
		screen.CalendarNext = artworkPickerLink("months", month, year+1, shownDate)
		screen.CalendarCells = artworkCalendarMonths(year, createdOn, enabled, shownDate, screens.openedOn)
	case "years":
		screen.CalendarYearLabel = strconv.Itoa(blockStart) + "–" + strconv.Itoa(blockStart+11)
		screen.CalendarPrev = artworkPickerLink("years", month, blockStart-12, shownDate)
		screen.CalendarNext = artworkPickerLink("years", month, blockStart+12, shownDate)
		screen.CalendarCells = artworkCalendarYears(blockStart, createdOn, enabled, shownDate, screens.openedOn)
	default:
		screen.CalendarPrev = artworkPickerLink("days", month.AddDate(0, -1, 0), month.Year(), shownDate)
		screen.CalendarNext = artworkPickerLink("days", month.AddDate(0, 1, 0), month.Year(), shownDate)
		screen.CalendarCells = artworkCalendarDays(month, shownDate, screens.openedOn)
	}
	screens.writeArtworkScreen(w, "artworks-created-on-tile-screen", screen)
}

func (screens *ArtworkScreens) showArtworkFeed(w http.ResponseWriter, r *http.Request) {
	id := 0
	if raw := r.PathValue("id"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			screens.showMissingArtwork(w)
			return
		}
		id = parsed
	}
	artwork, ok := screens.artworkLookup.OpenArtworkFeed(id, r.URL.Query().Get("next") == "true")
	if !ok {
		screens.showMissingArtwork(w)
		return
	}
	mediaBase := screens.mediaBase()
	screens.writeArtworkScreen(w, "artworks-feed-screen", ArtworkScreenData{
		Title:           artwork.ArtworkTitle,
		ArtworkTab:      "artworks-feed",
		ArtworkFileBase: mediaBase,
		Artwork:         screens.artworkTileView(artwork, mediaBase),
	})
}

func (screens *ArtworkScreens) showArtworkDraftForm(w http.ResponseWriter, r *http.Request) {
	artwork, ok := screens.artworkLookup.OnlyArtworkDraft()
	if !ok {
		screens.showMissingArtwork(w)
		return
	}
	mediaBase := screens.mediaBase()
	screen := ArtworkScreenData{
		Title:                "Заявка на экспертизу",
		ArtworkTab:          "artworks-draft-form",
		ArtworkFileBase:     mediaBase,
		Artwork:             screens.artworkTileView(artwork, mediaBase),
		DraftArtworkTitle:   artwork.ArtworkTitle,
		DraftAttributionNote: artwork.AttributionNote,
		DraftCreatedOn:       artwork.CreatedOn.Format("02.01.2006"),
		DraftDiagonal:        strconv.Itoa(artwork.DiagonalCm),
	}
	query := r.URL.Query()
	if _, sent := query["artwork_title"]; sent {
		screen.DraftArtworkTitle = query.Get("artwork_title")
		screen.DraftAttributionNote = query.Get("attribution_note")
		screen.DraftCreatedOn = query.Get("created_on")
		screen.DraftDiagonal = query.Get("diagonal_cm")
	}
	screens.writeArtworkScreen(w, "artworks-draft-form-screen", screen)
}

func (screens *ArtworkScreens) showMissingArtwork(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	screens.writeArtworkScreen(w, "artworks-feed-screen", ArtworkScreenData{
		Title:          "Картина не найдена",
		ArtworkTab:     "artworks-feed",
		ArtworkMissing: true,
	})
}

func (screens *ArtworkScreens) writeArtworkScreen(w http.ResponseWriter, name string, screen ArtworkScreenData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := screens.artworkTmpl.ExecuteTemplate(w, name, screen); err != nil {
		http.Error(w, "шаблон картины не собран", http.StatusInternalServerError)
	}
}

func (screens *ArtworkScreens) artworkTileViews(published []model_artworks.ArtworkAttribution, mediaBase string) []ArtworkTileView {
	tiles := make([]ArtworkTileView, 0, len(published))
	for _, artwork := range published {
		tiles = append(tiles, screens.artworkTileView(artwork, mediaBase))
	}
	return tiles
}

var artworkStorageReachable = func(base string) bool {
	client := &http.Client{Timeout: 400 * time.Millisecond}
	response, err := client.Get(strings.TrimRight(base, "/") + "/minio/health/live")
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == http.StatusOK
}

func (screens *ArtworkScreens) mediaBase() string {
	root := strings.TrimRight(screens.artworkFileBase, "/")
	if root != "" && artworkStorageReachable(root) {
		return root + "/artworks"
	}
	return "/artworks-default"
}

var artworkMonthNames = [...]string{
	"", "Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
	"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь",
}

var artworkMonthShort = [...]string{
	"", "Янв", "Фев", "Мар", "Апр", "Май", "Июн",
	"Июл", "Авг", "Сен", "Окт", "Ноя", "Дек",
}

func artworkCalendarMonth(raw string, createdOn time.Time, enabled bool, openedOn time.Time) time.Time {
	if parsed, err := time.Parse("2006-01", raw); err == nil {
		return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	if enabled {
		return time.Date(createdOn.Year(), createdOn.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(openedOn.Year(), openedOn.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func artworkPickerLink(view string, month time.Time, year int, shownDate string) string {
	query := url.Values{}
	query.Set("calendar", view)
	if view == "days" {
		query.Set("month", month.Format("2006-01"))
	} else {
		query.Set("year", strconv.Itoa(year))
	}
	if shownDate != "" {
		query.Set("created_on", shownDate)
	}
	return "/artworks-created-on-tile?" + query.Encode()
}

func artworkChosenDateLink(day time.Time) string {
	query := url.Values{}
	query.Set("created_on", day.Format("02.01.2006"))
	return "/artworks-created-on-tile?" + query.Encode()
}

func artworkCalendarDays(month time.Time, shownDate string, openedOn time.Time) []ArtworkCalendarCell {
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	lead := (int(first.Weekday()) + 6) % 7
	lastDay := time.Date(month.Year(), month.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	days := make([]ArtworkCalendarCell, 0, lead+lastDay)
	for i := 0; i < lead; i++ {
		days = append(days, ArtworkCalendarCell{})
	}
	openedValue := openedOn.Format("02.01.2006")
	for day := 1; day <= lastDay; day++ {
		date := time.Date(month.Year(), month.Month(), day, 0, 0, 0, 0, time.UTC)
		value := date.Format("02.01.2006")
		days = append(days, ArtworkCalendarCell{
			Label:  strconv.Itoa(day),
			Href:   artworkChosenDateLink(date),
			Chosen: value == shownDate || (shownDate == "" && value == openedValue),
		})
	}
	return days
}

func artworkCalendarMonths(year int, createdOn time.Time, enabled bool, shownDate string, openedOn time.Time) []ArtworkCalendarCell {
	months := make([]ArtworkCalendarCell, 0, 12)
	for month := time.January; month <= time.December; month++ {
		date := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		months = append(months, ArtworkCalendarCell{
			Label:  artworkMonthShort[month],
			Href:   artworkPickerLink("days", date, year, shownDate),
			Chosen: (enabled && createdOn.Year() == year && createdOn.Month() == month) || (!enabled && openedOn.Year() == year && openedOn.Month() == month),
		})
	}
	return months
}

func artworkCalendarYears(blockStart int, createdOn time.Time, enabled bool, shownDate string, openedOn time.Time) []ArtworkCalendarCell {
	years := make([]ArtworkCalendarCell, 0, 12)
	for year := blockStart; year < blockStart+12; year++ {
		years = append(years, ArtworkCalendarCell{
			Label:  strconv.Itoa(year),
			Href:   artworkPickerLink("months", time.Time{}, year, shownDate),
			Chosen: (enabled && createdOn.Year() == year) || (!enabled && openedOn.Year() == year),
		})
	}
	return years
}

func parseArtworkDate(raw string) (time.Time, bool) {
	for _, layout := range []string{"2006-01-02", "02.01.2006"} {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func (screens *ArtworkScreens) artworkTileView(artwork model_artworks.ArtworkAttribution, mediaBase string) ArtworkTileView {
	feedURL := "/artworks/" + strconv.Itoa(artwork.ArtworkID)
	return ArtworkTileView{
		ArtworkTitle:        artwork.ArtworkTitle,
		AttributionNote:      artwork.AttributionNote,
		CreatedOn:            artwork.CreatedOn.Format("02.01.2006"),
		DiagonalCm:           artwork.DiagonalCm,
		PainterName:          artwork.PainterName,
		ArtworkFileBase:     mediaBase,
		ArtworkImageKey:     artwork.ArtworkImageKey,
		ArtworkVideoKey:     artwork.ArtworkVideoKey,
		ConnoisseurMarkCount: len(artwork.MarkConnoisseurIDs),
		ArtworkFeedURL:      feedURL,
		NextArtworkFeedURL:  feedURL + "?next=true",
	}
}
