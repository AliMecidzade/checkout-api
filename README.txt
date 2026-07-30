1. GET /users/{id}/orders — Order History
# success: get all orders for user 1
curl.exe http://localhost:8080/users/1/orders

# fail: invalid user ID → 400
curl.exe -v http://localhost:8080/users/abc/orders
2. POST /orders — Create Order
# success: valid order with laptop × 1
curl.exe -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{"user_id":1,"items":[{"item_id":1,"quantity":1}]}'

# fail: user_id = 0 → 400
curl.exe -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{"user_id":0,"items":[{"item_id":1,"quantity":1}]}' -v

# fail: empty items → 400
curl.exe -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{"user_id":1,"items":[]}' -v

# fail: item not found → 400
curl.exe -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{"user_id":1,"items":[{"item_id":999,"quantity":1}]}' -v

# fail: quantity = 0 → 400
curl.exe -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{"user_id":1,"items":[{"item_id":1,"quantity":0}]}' -v

# fail: invalid JSON body → 400
curl.exe -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{invalid}' -v

# fail: wrong method → 405
curl.exe -v http://localhost:8080/orders
3. GET /items — Query Parameters & Caching
# success: all items
curl.exe http://localhost:8080/items

# success: filter by category
curl.exe "http://localhost:8080/items?category=Peripherals"

# success: sort by price
curl.exe "http://localhost:8080/items?sort=price"

# success: category + sort combined
curl.exe "http://localhost:8080/items?category=Electronics&sort=price"

# fail: category not found → 404
curl.exe -v "http://localhost:8080/items?category=Foo"

# fail: invalid sort field → 400
curl.exe -v "http://localhost:8080/items?sort=invalid"
4. GET /items/{id} — Single Item
# success: get item by ID
curl.exe http://localhost:8080/items/1

# fail: item not found → 404
curl.exe -v http://localhost:8080/items/999

# fail: invalid ID format → 400
curl.exe -v http://localhost:8080/items/abc