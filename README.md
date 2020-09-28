# Mercado Libre - Challenge
[![Build Status](https://travis-ci.com/jmanzanog/meli-challenge.svg?branch=master)](https://travis-ci.com/jmanzanog/meli-challenge)
## Description

Se pide construir una API  Go que cumpla con los siguientes requerimientos:

#### Request que debe responder:
```batch
curl -X GET 'http://localhost:4500/items/$ITEM_ID'
```
#### Respuesta
Debe responder la informacion unificada consultando a las siguientes APIs:
1. Api: https://api.mercadolibre.com/items/$ITEM_ID
   Descripcion: Informacion del Item
2. Api: https://api.mercadolibre.com/items/$ITEM_ID/children
   Descripcion: Informacion de los items hijos

## Data Base Design
![](https://github.com/jmanzanog/meli-challenge/blob/master/DiagramaER.png)

### Build desde Docker file
```batch
docker build . -t meli-challenge
docker run -itd -p 4500:4500 meli-challenge
```



### Correr desde Docker Hub 
```batch
docker run -itd -p 4500:4500  jmanzanog/meli-challenge
```

### Reponse JSON
```batch
curl --location --request GET 'http://localhost:4500/items/MLU460998489'
{
   "item_id":"MLU460998489",
   "title":"Google Pixel 32gb Silver - Impecable!",
   "category_id":"MLU1055",
   "price":350.00,
   "start_time":"2019-03-02T20:31:02.000+0000",
   "stop_time":"2019-10-25T23:28:35.000+0000",
   "children":[
      {
         "item_id":"MLU468887129",
         "stop_time":"2020-04-25T22:10:52.000+0000"
      }
   ]
}
```   
```batch
curl --location --request GET 'http://localhost:4500/health'
[
 {
    "date": "2020-09-16T00:28:00.000261878Z",
    "avg_response_time": 0.00000175173704192054,
    "total_requests": 1028,
    "avg_response_time_api_calls":  0.01020017517375419278,
    "total_count_api_calls": 1,
    "info_requests": [
      {
        "status_code": 200,
        "count": 1025
      }
    ]
  },
  {
    "date": "2020-09-16T00:29:00.013205887Z",
    "avg_response_time": 0.000007584508156045672,
    "total_requests": 692,
    "avg_response_time_api_calls": 0,
    "total_count_api_calls": 0,
    "info_requests": [
      {
        "status_code": 200,
        "count": 692
      }
    ]
  },
  {
    "date": "2020-09-16T00:30:00.00032474Z",
    "avg_response_time": 0.000012031942939073698,
    "total_requests": 211,
    "avg_response_time_api_calls": 0,
    "total_count_api_calls": 0,
    "info_requests": [
      {
        "status_code": 200,
        "count": 211
      }
    ]
  },
  {
    "date": "2020-09-16T00:31:00.000269661Z",
    "avg_response_time": 0,
    "total_requests": 0,
    "avg_response_time_api_calls": 0,
    "total_count_api_calls": 0,
    "info_requests": null
  },
]
``` 

### Comentarios  Mejoras
* CI  agregar jobs al travis CI file para agregar continus integration (unit test, integration test, lint tools (SonarQube)).
* CD el proyecto utiliza travis CI para construir la imagen de docker y subirla al Docker Hub pero no la está desplegando en un ambiente productivo (continus delivery) así que se podría agregar este Job
* La aplicación usa postgresql database para hacer un cache de la información yo cambiaria postgrest por REDIS o MEMCACHED database que están optimizadas para resolver estos problemas
* La aplicación ya esta dockerizada así que el paso siguiente es usar kubernetes para escalarla horizontalmente 

 
 
