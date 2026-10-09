---
name: exhibition-database
description: Проектирует PostgreSQL и подключает каталог произведений для лабораторной 2 курса iu5git/Web. Use when adding the exhibition schema, GORM models, Adminer seed data, publishing an artwork, soft-withdraw via SQL UPDATE, or wiring the three lab 1 templates to the database.
---

# База каталога, лабораторная 2

Читать `документация/лабораторная-2.md` и `документация/модель-данных.md`.

## Таблицы

Три таблицы, без каскадного удаления:

- `artworks` — название, этикетка, `display_status`, ключи изображения и видео, `creation_year`, `gallery_hall`, `catalogued_at`, `curator_id`, `label_formed_at`
- `visitors` — посетители и кураторы
- `visitor_appreciations` — м-м, свой PK и два FK

В данных сразу три произведения: черновик, в экспозиции, снято. У посетителя не больше одного черновика.

## Методы

Шесть HTTP-методов:

1. GET лента — одна строка из БД
2. GET черновик
3. GET плитка с фильтром `creation_year`
4. POST «Далее» — создать черновик через GORM, в ключи записать заведомо неверные URL (файлы на сервер не слать)
5. POST «Опубликовать» — статус `on_display` через GORM после названия, этикетки, года и зала
6. POST снятие с показа — `UPDATE` через `database/sql`, не через GORM

Отметки только читать из БД, не ставить. Снятое произведение по URL не открывать. Пустой или битый ключ медиа — картинка и ролик по умолчанию из `static/`.

## Проверка

Adminer: смена полей темы и числа отметок видна на страницах. `select` до и после публикации и снятия. В коде пять обращений через ORM и один SQL `UPDATE`.
