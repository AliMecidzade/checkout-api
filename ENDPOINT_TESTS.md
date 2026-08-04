# Checkout API — тесты эндпоинтов через curl.exe

Порт из `.env`: **8090** (если изменился, замени на свой).

> **ОБЯЗАТЕЛЬНО:** PowerShell 5.1 некорректно передаёт аргументы с фигурными скобками/кавычками
> в native-команды (тело обрезается → `invalid request body`). Надёжный способ — подавать JSON
> через **стандартный ввод (stdin)**:
> ```powershell
> $json = '{"user_id":1,"quantity":5}'      # одиночные кавычки
> $json | curl.exe -d "@-"                  # передача тела через stdin
> ```
> Альтернатива (если не работает): `-d @путь_к_файлу`, записав JSON в файл.

---

## 1. Корзина пользователя 1

```powershell
# 1.1 Создать корзину + добавить товары -> 201
$cart = '{"user_id":1,"items":[{"item_id":1,"quantity":2},{"item_id":3,"quantity":1}]}'
$cart | curl.exe -i -X POST http://localhost:8090/user/cart -H "Content-Type: application/json" -d "@-"

# 1.2 Прочитать корзину -> 200
curl.exe http://localhost:8090/user/cart -H "X-User-ID: 1"

# 1.3 Повторное создание той же корзины -> 409 Conflict
$dup = '{"user_id":1,"items":[{"item_id":2,"quantity":1}]}'
$dup | curl.exe -i -X POST http://localhost:8090/user/cart -H "Content-Type: application/json" -d "@-"

# 1.4 Обновить кол-во item_id=1 -> 200
$upd = '{"user_id":1,"quantity":5}'
$upd | curl.exe -i -X PATCH http://localhost:8090/user/cart/items/1 -H "Content-Type: application/json" -d "@-"

# 1.5 Удалить item_id=3 -> 204
$del = '{"user_id":1}'
$del | curl.exe -i -X DELETE http://localhost:8090/user/cart/items/3 -H "Content-Type: application/json" -d "@-"

# 1.6 Прочитать корзину после удаления -> 200
curl.exe http://localhost:8090/user/cart -H "X-User-ID: 1"
```

---

## 2. Заказ из корзины

```powershell
# 2.1 Создать заказ -> 201 (при успешной оплате корзина удаляется)
$order = '{"user_id":1}'
$order | curl.exe -i -X POST http://localhost:8090/user/orders -H "Content-Type: application/json" -H "Idempotency-Key: test-1" -d "@-"

# 2.2 Повтор с тем же ключом -> тот же ответ из кэша (идемпотентность)
$order | curl.exe -i -X POST http://localhost:8090/user/orders -H "Content-Type: application/json" -H "Idempotency-Key: test-1" -d "@-"
```

> Примечание: шаг 2.1 удаляет корзину юзера 1. Для повторного прогона пересоздай корзину
> (раздел 1) или используй юзера 2/3 (в коде ограничений на user_id нет).

---

## 3. Товары

```powershell
# 3.1 Все товары -> 200
curl.exe http://localhost:8090/items

# 3.2 Товар по ID -> 200
curl.exe http://localhost:8090/items/1

# 3.3 Несуществующий ID -> 404
curl.exe -i http://localhost:8090/items/999

# 3.4 Невалидный ID -> 400
curl.exe -i http://localhost:8090/items/abc
```

---

## 4. Ожидаемые ошибки (проверка валидаций)

```powershell
# 4.1 GET /user/cart без X-User-ID -> 400
curl.exe -i http://localhost:8090/user/cart

# 4.2 PATCH с quantity <= 0 -> 400
$bad = '{"user_id":1,"quantity":0}'
$bad | curl.exe -i -X PATCH http://localhost:8090/user/cart/items/1 -H "Content-Type: application/json" -d "@-"

# 4.3 Создание заказа без Idempotency-Key -> 400
curl.exe -i -X POST http://localhost:8090/user/orders -H "Content-Type: application/json" -d "@-"

# 4.4 Метод не разрешён (POST на GET-эндпоинт /items) -> 405
curl.exe -i -X POST http://localhost:8090/items

# 4.5 Пустая корзина при создании -> 400
$empty = '{"user_id":2,"items":[]}'
$empty | curl.exe -i -X POST http://localhost:8090/user/cart -H "Content-Type: application/json" -d "@-"
```

---

## Полный сценарий (по порядку)

1. **1.1** — создать корзину юзера 1
2. **1.2** — прочитать её
3. **1.4** — обновить количество
4. **1.5** — удалить один товар
5. **1.6** — прочитать после удаления
6. **2.1** — заказ из корзины (оплата → удаление корзины)
7. **2.2** — повтор с тем же Idempotency-Key
8. **3.x / 4.x** — по желанию

## Ссылка на точки входа

Полный список маршрутов в `main.go:35-60`:

| Метод | Путь | Handler |
|---|---|---|
| GET | `/items` | GetItems |
| GET | `/items/{id}` | GetItemByID |
| POST | `/user/cart` | CreateUserCartAndAddItems |
| GET | `/user/cart` | GetUserCart |
| PATCH | `/user/cart/items/{item_id}` | UpdateCartItem |
| DELETE | `/user/cart/items/{item_id}` | RemoveCartItem |
| POST | `/user/orders` | CreateOrderFromCart |