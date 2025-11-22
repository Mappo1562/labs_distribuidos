Laboratorio 3: Sistemas Distribuidos
=============================================
--------------------------------------------------------------------
AMBIENTE DE EJECUCION Y DESARROLLO 
--------------------------------------------------------------------
- Ejecucion: Docker desktop 28.3.2
- Desarollo: Visual Studio Code

--------------------------------------------------------------------
EJECUCIÓN
--------------------------------------------------------------------

- si se quiere ejecutar en las VM, hay que ejecutar todo en la 
  carpeta VMs, es importante comenzar siempre por la VM dist101 ya 
  que ahi esta el broker.
  Los comandos deben ser ejecutados en orden.
dist042:
    sudo make build-VM1
    sudo make down-VM1
    sudo make docker-VM1

dist043:
    sudo make build-VM2
    sudo make down-VM2
    sudo make docker-VM2

dist044:
    sudo make build-VM3
    sudo make down-VM3
    sudo make docker-VM3

dist101:
    sudo make build-VM4
    sudo make down-VM4
    sudo make docker-VM4

--------------------------------------------------------------------
CONSIDERACIONES AL MOMENTO DE EJECUTAR
--------------------------------------------------------------------
  
--------------------------------------------------------------------
NOMBRE Y ROL INTEGRANTES
--------------------------------------------------------------------
- Emiliano Garcia 
- 202273622-4

- Alejandro Sánchez
- 202273605-4

- Alejandro Vergara
- 202273568-6