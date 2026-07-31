1.CREATE A CART WITH SOME ITEMS (paste again for cart already created error)
curl -i -X POST http://localhost:8080/user/cart \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "items": [
      {
        "item_id": 1,
        "quantity": 1
      },
      {
        "item_id": 2,
        "quantity": 2
      },
      {
        "item_id": 3,
        "quantity": 1
      }
    ]
  }'
  
  
  
2. GET THE CARD BY USER ID IN THE HEADER 
curl -i http://localhost:8080/user/cart \
  -H "X-User-ID: 1"
  

3. GET CART WITHOUT USER ID 
curl -i http://localhost:8080/user/cart

4. CREATING EMPTY CART 
curl -i -X POST http://localhost:8080/user/cart \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "items": []
  }'
 
 
5. CREATE EMPTY CART 
curl -i -X POST http://localhost:8080/user/cart \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "items": []
  }'
  
6. CREATING CART WITH NON-EXISTING ITEM 
curl -i -X POST http://localhost:8080/user/cart \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "items": [
      {
        "item_id": 999,
        "quantity": 1
      }
    ]
  }'
  
  
7. PATCH CHANGE THR QUANTITY 
curl -i -X PATCH http://localhost:8080/user/cart/items/2 \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "quantity": 5
  }'
  
  
8. PATCH WITH 0 QUANTITY
curl -i -X PATCH http://localhost:8080/user/cart/items/2 \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "quantity": 0
  }'
  
9. PATCH WITH QUANTITY < 0 

curl -i -X PATCH http://localhost:8080/user/cart/items/2 \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "quantity": -19
  }'

10. PATCH NON-Existing ITEM 
curl -i -X PATCH http://localhost:8080/user/cart/items/999 \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "quantity": 3
  }'
  
11. DELETE ITEM FROM CART 
curl -i -X DELETE http://localhost:8080/user/cart/items/2 \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1
  }'

curl -s http://localhost:8080/user/cart \
  -H "X-User-ID: 1" | jq
  
14. DELETE LAST ITEM 
curl -i -X DELETE http://localhost:8080/user/cart/items/1 \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1
  }'
  
curl -i -X DELETE http://localhost:8080/user/cart/items/3 \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1
  }'
  
  
  
curl -s http://localhost:8080/user/cart \
  -H "X-User-ID: 1" | jq
  
17. CREATE CART FOR PLACE ORDER 
curl -i -X POST http://localhost:8080/user/cart \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 3,
    "items": [
      {
        "item_id": 1,
        "quantity": 1
      },
      {
        "item_id": 2,
        "quantity": 2
      }
    ]
  }'


18. PLACE ORDER WITHOUT IDEMPOTENCY KEY

curl -i -X POST http://localhost:8080/user/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 3
  }'


19. SUCCESFULLY PLACE ORDER 
curl -i -X POST http://localhost:8080/user/orders \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: order-user-3-001" \
  -d '{
    "user_id": 3
  }'
  
AFTER ORDER PLACING CART DELETES 
curl -s http://localhost:8080/user/cart \
  -H "X-User-ID: 3" | jq   
  
20. PLACE ORDER WITH THE SAME KEY 
curl -i -X POST http://localhost:8080/user/orders \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: order-user-3-001" \
  -d '{
    "user_id": 3
  }'
  
21. USE OLD KEY BUT NEW USER 
curl -i -X POST http://localhost:8080/user/orders \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: order-user-3-001" \
  -d '{
    "user_id": 4
  }'
  
