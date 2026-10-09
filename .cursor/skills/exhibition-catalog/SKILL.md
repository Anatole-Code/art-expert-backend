---
name: exhibition-catalog
description: Собирает страницы каталога художественных выставок на Go и html/template по лабораторной 1 курса iu5git/Web и разметке Prado из .example. Use when implementing lab 1, the home page, explore-artworks, an artwork page, an artist page, the draft wall-label page, MinIO image keys, or the exhibition menu.
---

# Каталог выставки, лабораторная 1

Читать `документация/лабораторная-1.md` и `документация/стиль-prado.md` до правок шаблонов.

## Страницы

| Страница | URL | Выборка |
| --- | --- | --- |
| Главная, меню | `/` | опубликованные, кратко |
| Коллекция (плитка) | `/explore-artworks?creation_year=` | `on_display`, фильтр по году |
| Произведение (лента) | `/the-collection/art-work/{slug}/{id}?next=true` | одна работа; `next` — следующая |
| Автор | `/artist/{slug}/` | работы этого автора в экспозиции |
| Черновик этикетки | `/curator/wall-label` | единственный `in_preparation`, без сохранения |

Клик по карточке плитки ведёт на ленту с этим id. Снятые с показа не рендерить. У куратора не больше одного черновика.

## Коллекция

Один срез `Artwork` в памяти. Вложенный список id посетителей допустим только для отметок. `creation_year` и `gallery_hall` — скаляры. Счётчик отметок считает обработчик.

Образец: El Greco, *The Nobleman with his Hand on his Chest*, id `9`, slug `the-nobleman-with-his-hand-on-his-chest`, автор `el-greco-domenikos-theotokopoulos`. Плюс черновик и снятая работа.

## MinIO

Поднимать `docker-compose.yml` (порты 9000/9001). В модели два ключа латиницей; в HTML — полные URL бакета. Ключи показать в трёх шаблонах: лента, плитка, черновик.

## Проверка

- Три GET курса видны во вкладке Network: лента, черновик, плитка с `creation_year`.
- В Response плитки есть URL картинки и видео из MinIO.
- Поиск года остаётся в поле после запроса.
- JavaScript на страницах нет.
