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
	previousFile := artworkFileReachable
	artworkStorageReachable = func(string) bool { return true }
	artworkFileReachable = func(string) bool { return true }
	t.Cleanup(func() {
		artworkStorageReachable = previousReachable
		artworkFileReachable = previousFile
	})
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
		{"/artworks-created-on-tile?created_on=1580-12-31", []string{`Дата создания (не позднее)`, `placeholder="дд.мм.гггг"`, `value="31.12.1580"`, "Декабрь", "1580", ">пн<", "Показать", "Удалить", "nobleman-hand-chest.jpg", "01.06.1580", "81 см", "http://127.0.0.1:9000/artworks/"}, []string{"The Three Graces", "The Virgin of the Chair", "The Virgin Mary", "<script", `type="date"`}, 200},
		{"/artworks-created-on-tile?calendar=months&year=1580&created_on=31.12.1580", []string{"Янв", "Дек", "open", "31.12.1580"}, []string{"<script"}, 200},
		{"/artworks-created-on-tile?calendar=years&year=1572&created_on=31.12.1580", []string{"1572–1583", "1580", "open"}, []string{"<script"}, 200},
		{"/artworks-draft-form", []string{"virgin-of-the-chair.jpg", "Опубликовать", `action="/artworks-draft-form/publish"`}, []string{"запись не выполняется", "Отправить заявку"}, 200},
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
	if !strings.Contains(body, defaultArtworkImageSrc) || !strings.Contains(body, defaultArtworkVideoSrc) || strings.Contains(body, "127.0.0.1:9000") {
		t.Fatalf("без хранилища лента не взяла файлы по умолчанию: %s", body)
	}
	artworkStorageReachable = func(string) bool { return true }
	artworkFileReachable = func(fileURL string) bool { return !strings.Contains(fileURL, "missing-artwork") }
	lookup := model_artworks.NewMemoryArtworkLookup()
	if _, err := lookup.PublishArtworkDraft("Краткое описание", time.Date(1580, time.June, 1, 0, 0, 0, 0, time.UTC), 40); err != nil {
		t.Fatal(err)
	}
	draft, err := lookup.CreateArtworkDraft("Новая картина")
	if err != nil {
		t.Fatal(err)
	}
	missingScreens := OpenArtworkScreens(lookup, "http://127.0.0.1:9000", "", openedOn)
	response = httptest.NewRecorder()
	missingScreens.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/artworks-draft-form", nil))
	body = response.Body.String()
	if !strings.Contains(body, defaultArtworkImageSrc) || !strings.Contains(body, defaultArtworkVideoSrc) || strings.Contains(body, "missing-artwork") {
		t.Fatalf("недоступный файл черновика не заменён файлом по умолчанию: %s", body)
	}
	if _, err := lookup.PublishArtworkDraft("Описание новой", time.Date(1600, time.January, 2, 0, 0, 0, 0, time.UTC), 33); err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	missingScreens.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/artworks/"+strconv.Itoa(draft.ArtworkID), nil))
	body = response.Body.String()
	if !strings.Contains(body, defaultArtworkImageSrc) || !strings.Contains(body, defaultArtworkVideoSrc) {
		t.Fatalf("лента новой картины не взяла файлы по умолчанию: %s", body)
	}
	response = httptest.NewRecorder()
	missingScreens.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/artworks-created-on-tile", nil))
	body = response.Body.String()
	if !strings.Contains(body, "http://127.0.0.1:9000/artworks/nobleman-hand-chest.jpg") || !strings.Contains(body, defaultArtworkImageSrc) {
		t.Fatalf("плитка не различила доступный файл и файл по умолчанию: %s", body)
	}
}

func TestPublishCreateAndRemove(t *testing.T) {
	lookup := model_artworks.NewMemoryArtworkLookup()
	artworkScreens := OpenArtworkScreens(lookup, "http://127.0.0.1:9000", "", time.Now())

	publish := httptest.NewRequest(http.MethodPost, "/artworks-draft-form/publish", strings.NewReader("attribution_note=Краткое+описание&created_on=01.06.1580&diagonal_cm=40"))
	publish.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	artworkScreens.ServeHTTP(response, publish)
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/artworks/21" {
		t.Fatalf("publish %d %s", response.Code, response.Header().Get("Location"))
	}
	if _, ok := lookup.OnlyArtworkDraft(); ok {
		t.Fatal("draft remained after publish")
	}

	create := httptest.NewRequest(http.MethodPost, "/artworks-draft-form/next", strings.NewReader("artwork_title=Новая+картина"))
	create.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	artworkScreens.ServeHTTP(response, create)
	draft, ok := lookup.OnlyArtworkDraft()
	if !ok || draft.ArtworkTitle != "Новая картина" || draft.ArtworkImageKey != "missing-artwork-image" || draft.ArtworkVideoKey != "missing-artwork-video" {
		t.Fatalf("draft %+v ok=%v", draft, ok)
	}

	remove := httptest.NewRequest(http.MethodPost, "/artworks/9/remove", strings.NewReader("created_on=31.12.1580"))
	remove.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	artworkScreens.ServeHTTP(response, remove)
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/artworks-created-on-tile?created_on=31.12.1580" {
		t.Fatalf("remove %d %s", response.Code, response.Header().Get("Location"))
	}
	if _, ok := lookup.OpenArtworkFeed(9, false); ok {
		t.Fatal("removed artwork stayed in the feed")
	}
	if err := lookup.RemovePublishedArtwork(draft.ArtworkID); err == nil {
		t.Fatal("draft was removed by the published-only method")
	}
}
