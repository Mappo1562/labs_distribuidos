Laboratorio 1: Sistemas Distribuidos
=============================================
--------------------------------------------------------------------
AMBIENTE DE EJECUCION Y DESARROLLO 
--------------------------------------------------------------------
- Ejecucion: Docker desktop 28.3.2
- Desarollo: Visual Studio Code

--------------------------------------------------------------------
EJECUCIÓN
--------------------------------------------------------------------
- en la carpeta lab 1 hay que ejecutar (con Docker desktop abierto):
    docker-compose up --build
para ejecutarlo en una sola maquina

- si se quiere ejecutar en las VM, hay que seguir el siguiente orden
ubicacion (labs_distribuidos/lab1/vm/<dist___>)
dist042:
    navegar hasta la carpeta Dist042 y ejecutar:
        make docker-lester
dist044:
    navegar hasta la carpeta Dist044 y ejecutar:
        make docker-franklin
dist101:
    navegar hasta la carpeta Dist101 y ejecutar:
        make docker-trevor
dist043:
    navegar hasta la carpeta Dist043 y ejecutar:
        make docker-michael
si alguno tira error, ejecutar:
    make docker-restart

--------------------------------------------------------------------
CONSIDERACIONES AL MOMENTO DE EJECUTAR
--------------------------------------------------------------------
- Se necesita que el archivo que lea lester exista

--------------------------------------------------------------------
NOMBRE Y ROL INTEGRANTES
--------------------------------------------------------------------
- Emiliano Garcia 
- 202273622-4

- Alejandro Sánchez
- 202273605-4

- Alejandro Vergara
- 202273568-6