package controller_artworkscreens

import (
	"errors"
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
	ArtworkTitle         string
	AttributionNote      string
	CreatedOn            string
	DiagonalCm           int
	PainterName          string
	ArtworkFileBase      string
	ArtworkImageKey      string
	ArtworkVideoKey      string
	ArtworkImageSrc      string
	ArtworkVideoSrc      string
	ConnoisseurMarkCount int
	ArtworkFeedURL       string
	NextArtworkFeedURL   string
	ArtworkID            int
	FilterCreatedOn      string
}

type ArtworkCalendarCell struct {
	Label  string
	Href   string
	Chosen bool
}

type ArtworkScreenData struct {
	Title                string
	ArtworkTab           string
	ArtworkFileBase      string
	ArtworkTiles         []ArtworkTileView
	Artwork              ArtworkTileView
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
	DraftArtworkTitle    string
	DraftAttributionNote string
	DraftCreatedOn       string
	DraftDiagonal        string
	DraftFormAction      string
	DraftSubmitLabel     string
	DraftFormError       string
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
		ArtworkTiles:      screens.artworkTileViews(published, mediaBase, shownDate),
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
	artwork, ready := screens.artworkLookup.OnlyArtworkDraft()
	screen := screens.draftForm(artwork, ready)
	query := r.URL.Query()
	if _, sent := query["artwork_title"]; sent {
		screen.DraftArtworkTitle = query.Get("artwork_title")
		screen.DraftAttributionNote = query.Get("attribution_note")
		screen.DraftCreatedOn = query.Get("created_on")
		screen.DraftDiagonal = query.Get("diagonal_cm")
	}
	screens.writeArtworkScreen(w, "artworks-draft-form-screen", screen)
}

func (screens *ArtworkScreens) createArtworkDraft(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "форма черновика не прочитана", http.StatusBadRequest)
		return
	}
	title := strings.TrimSpace(r.FormValue("artwork_title"))
	_, err := screens.artworkLookup.CreateArtworkDraft(title)
	if err == nil {
		http.Redirect(w, r, "/artworks-draft-form", http.StatusSeeOther)
		return
	}
	if errors.Is(err, model_artworks.ErrArtworkDraftExists) {
		http.Redirect(w, r, "/artworks-draft-form", http.StatusSeeOther)
		return
	}
	screen := screens.draftForm(model_artworks.ArtworkAttribution{}, false)
	screen.DraftArtworkTitle = title
	screen.DraftFormError = "Укажите название картины."
	w.WriteHeader(http.StatusBadRequest)
	screens.writeArtworkScreen(w, "artworks-draft-form-screen", screen)
}

func (screens *ArtworkScreens) publishArtworkDraft(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "форма публикации не прочитана", http.StatusBadRequest)
		return
	}
	note := strings.TrimSpace(r.FormValue("attribution_note"))
	createdOn, dateOK := parseArtworkDate(r.FormValue("created_on"))
	diagonal, diagonalErr := strconv.Atoi(strings.TrimSpace(r.FormValue("diagonal_cm")))
	draft, ready := screens.artworkLookup.OnlyArtworkDraft()
	if !ready {
		http.Redirect(w, r, "/artworks-draft-form", http.StatusSeeOther)
		return
	}
	if note == "" || !dateOK || diagonalErr != nil || diagonal <= 0 {
		screen := screens.draftForm(draft, true)
		screen.DraftAttributionNote = r.FormValue("attribution_note")
		screen.DraftCreatedOn = r.FormValue("created_on")
		screen.DraftDiagonal = r.FormValue("diagonal_cm")
		screen.DraftFormError = "Для публикации нужны описание, дата написания и диагональ."
		w.WriteHeader(http.StatusBadRequest)
		screens.writeArtworkScreen(w, "artworks-draft-form-screen", screen)
		return
	}
	published, err := screens.artworkLookup.PublishArtworkDraft(note, createdOn, diagonal)
	if err != nil {
		http.Redirect(w, r, "/artworks-draft-form", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/artworks/"+strconv.Itoa(published.ArtworkID), http.StatusSeeOther)
}

func (screens *ArtworkScreens) removeArtwork(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "форма удаления не прочитана", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		screens.showMissingArtwork(w)
		return
	}
	if err := screens.artworkLookup.RemovePublishedArtwork(id); err != nil {
		screens.showMissingArtwork(w)
		return
	}
	target := "/artworks-created-on-tile"
	if createdOn := strings.TrimSpace(r.FormValue("created_on")); createdOn != "" {
		target += "?created_on=" + url.QueryEscape(createdOn)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (screens *ArtworkScreens) draftForm(artwork model_artworks.ArtworkAttribution, ready bool) ArtworkScreenData {
	mediaBase := screens.mediaBase()
	screen := ArtworkScreenData{
		Title:            "Заявка на экспертизу",
		ArtworkTab:       "artworks-draft-form",
		ArtworkFileBase:  mediaBase,
		DraftFormAction:  "/artworks-draft-form/next",
		DraftSubmitLabel: "Далее",
	}
	if !ready {
		screen.Artwork.ArtworkImageSrc = defaultArtworkImageSrc
		screen.Artwork.ArtworkVideoSrc = defaultArtworkVideoSrc
		return screen
	}
	screen.Artwork = screens.artworkTileView(artwork, mediaBase)
	screen.DraftArtworkTitle = artwork.ArtworkTitle
	screen.DraftAttributionNote = artwork.AttributionNote
	if !artwork.CreatedOn.IsZero() {
		screen.DraftCreatedOn = artwork.CreatedOn.Format("02.01.2006")
	}
	if artwork.DiagonalCm != 0 {
		screen.DraftDiagonal = strconv.Itoa(artwork.DiagonalCm)
	}
	screen.DraftFormAction = "/artworks-draft-form/publish"
	screen.DraftSubmitLabel = "Опубликовать"
	return screen
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

func (screens *ArtworkScreens) artworkTileViews(published []model_artworks.ArtworkAttribution, mediaBase, shownDate string) []ArtworkTileView {
	tiles := make([]ArtworkTileView, 0, len(published))
	for _, artwork := range published {
		tile := screens.artworkTileView(artwork, mediaBase)
		tile.FilterCreatedOn = shownDate
		tiles = append(tiles, tile)
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

const (
	defaultArtworkImageSrc = "/artworks-default/default-artwork.jpg"
	defaultArtworkVideoSrc = "/artworks-default/default-artwork.mp4"
)

func (screens *ArtworkScreens) mediaBase() string {
	root := strings.TrimRight(screens.artworkFileBase, "/")
	if root != "" && artworkStorageReachable(root) {
		return root + "/artworks"
	}
	return "/artworks-default"
}

var artworkFileReachable = func(fileURL string) bool {
	request, err := http.NewRequest(http.MethodHead, fileURL, nil)
	if err != nil {
		return false
	}
	client := &http.Client{Timeout: 400 * time.Millisecond}
	response, err := client.Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == http.StatusOK
}

func (screens *ArtworkScreens) artworkMediaSrc(key, fallback string) string {
	key = strings.TrimSpace(key)
	base := screens.mediaBase()
	if key == "" || !strings.HasPrefix(base, "http") {
		return fallback
	}
	fileURL := base + "/" + key
	if !artworkFileReachable(fileURL) {
		return fallback
	}
	return fileURL
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
		ArtworkTitle:         artwork.ArtworkTitle,
		AttributionNote:      artwork.AttributionNote,
		CreatedOn:            artwork.CreatedOn.Format("02.01.2006"),
		DiagonalCm:           artwork.DiagonalCm,
		PainterName:          artwork.PainterName,
		ArtworkFileBase:      mediaBase,
		ArtworkImageKey:      artwork.ArtworkImageKey,
		ArtworkVideoKey:      artwork.ArtworkVideoKey,
		ArtworkImageSrc:      screens.artworkMediaSrc(artwork.ArtworkImageKey, defaultArtworkImageSrc),
		ArtworkVideoSrc:      screens.artworkMediaSrc(artwork.ArtworkVideoKey, defaultArtworkVideoSrc),
		ConnoisseurMarkCount: len(artwork.MarkConnoisseurIDs),
		ArtworkFeedURL:       feedURL,
		NextArtworkFeedURL:   feedURL + "?next=true",
		ArtworkID:            artwork.ArtworkID,
	}
}
