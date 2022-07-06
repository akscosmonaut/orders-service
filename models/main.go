GET orders-service/orders //список всех заказов

GET orders-service/orders/:id //список одного заказа
ответ:
{
'name':'Mariya', //имя клиента
'address': 'Тверская 6', //адрес дотавки
'phone':'+79268484744', //номер клиента
'status':'new',
'products': [1,6,8,4,3], //id продуктов из таблицы products
'total': 1480 //стоимость заказа (сумма цен пяти продуктов заказа)
}
POST orders-service/orders //создание заказа (заказ создаётся со статусом new)
body:
{
	'name':'Mariya', //имя клиента
	'address': 'Тверская 6', //адрес дотавки
	'phone':'+79268484744', //номер клиента
	'products': [1,6,8,4,3] //id продуктов из таблицы products
}
ответ: {'id': id, 'status': 'new'}
PUT orders-service/orders/:id //изменение состава заказа, контактной информации, адреса
body:
{
	'name':'Mariya',
	'address': 'Тверская 6, подъезд 9',
	'phone':'+79283834744',
	'products': [1,4,3]
}
ответ: 200

DELETE orders-service/orders/:id //удаление заказа
ответ: 200

GET orders-service/products //получение списка продуктов магазина
[{
	'id': 1,
	'name': 'хлеб',
	'price': 50
},
{
	'id': 2,
	'name': 'молоко',
	'price': 80
},
]

POST orders-service/change-status/:id //смена статуса заказа
body:
{'status': 'confirmed'}
или
{'status': 'canceled'}
или
{'status': 'done'}
//Возможные флоу заказов:
// new - confirmed - canceled
// new - confirmed - done
// new - canceled
