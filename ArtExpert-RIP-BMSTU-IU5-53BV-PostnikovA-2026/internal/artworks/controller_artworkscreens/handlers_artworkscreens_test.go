package controller_artworkscreens

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Anatole-Code/art-expert-backend/internal/artworks/model_artworks"
)

func TestArtworkScreens(t *testing.T) {
	previousReachable := artworkStorageReachable
	artworkStorageReachable = func(string) bool { return true }
	t.Cleanup(func() { artworkStorageReachable = previousReachable })
	openedOn := time.Now()
	artworkScreens := OpenArtworkScreens(model_artworks.NewMemoryArtworkLookup(), "http://127.0.0.1:9000", "", openedOn)
	checks := []struct {
		path string
		want []string
		ban  []string
		code int
	}{
		{"/artworks/9", []string{"Дата создания", "01.06.1580", "Диагональ, см", "81", "Следующая", "autoplay", "nobleman-hand-chest.mp4", "nobleman-hand-chest.jpg", "screen_artworks.css", "/artworks/"}, []string{"<script"}, 200},
		{"/artworks/9?next=true", []string{"The Virgin Mary"}, nil, 200},
		{"/artworks", []string{"The Nobleman with his Hand on his Chest"}, nil, 200},
		{"/artworks-created-on-tile", []string{`value="дд.мм.гггг"`, "Показать"}, []string{"<script"}, 200},
		{"/artworks-created-on-tile?created_on=1580-12-31", []string{`Дата создания (не позднее)`, `placeholder="дд.мм.гггг"`, `value="31.12.1580"`, "Декабрь", "1580", ">пн<", "Показать", "nobleman-hand-chest.jpg", "01.06.1580", "81 см", "http://127.0.0.1:9000/artworks/"}, []string{"The Three Graces", "The Virgin of the Chair", "The Virgin Mary", "<script", `type="date"`}, 200},
		{"/artworks-created-on-tile?calendar=months&year=1580&created_on=31.12.1580", []string{"Янв", "Дек", "open", "31.12.1580"}, []string{"<script"}, 200},
		{"/artworks-created-on-tile?calendar=years&year=1572&created_on=31.12.1580", []string{"1572–1583", "1580", "open"}, []string{"<script"}, 200},
		{"/artworks-draft-form", []string{"virgin-of-the-chair.jpg", "запись не выполняется"}, nil, 200},
		{"/artworks/30", []string{"не найдена", "/artworks-default/default-artwork.jpg", "/artworks-default/default-artwork.mp4"}, nil, 404},
	}
	checks[3].want = append(checks[3].want, artworkMonthNames[openedOn.Month()], strconv.Itoa(openedOn.Year()), ">"+strconv.Itoa(openedOn.Day())+"</a>")
	for _, check := range checks {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, check.path, nil)
		artworkScreens.ServeHTTP(response, request)
		body, _ := io.ReadAll(response.Body)
		artworkHTML := string(body)
		if response.Code != check.code {
			t.Fatalf("%s code %d body %s", check.path, response.Code, artworkHTML)
		}
		for _, want := range check.want {
			if !strings.Contains(artworkHTML, want) {
				t.Fatalf("%s missing %q", check.path, want)
			}
		}
		for _, ban := range check.ban {
			if strings.Contains(artworkHTML, ban) {
				t.Fatalf("%s contains %q", check.path, ban)
			}
		}
	}
	artworkStorageReachable = func(string) bool { return false }
	response := httptest.NewRecorder()
	artworkScreens.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/artworks/9", nil))
	body := response.Body.String()
	if !strings.Contains(body, "/artworks-default/nobleman-hand-chest.jpg") || strings.Contains(body, "127.0.0.1:9000") {
		t.Fatalf("без хранилища лента не взяла файл коллекции: %s", body)
	}
}
