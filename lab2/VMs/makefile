# Archivos docker-compose por VM
COMPOSE_VM1=Dist042/docker-compose.yml
COMPOSE_VM2=Dist043/docker-compose.yml
COMPOSE_VM3=Dist044/docker-compose.yml
COMPOSE_VM4=Dist101/docker-compose.yml

# ========== VM1 ==========
build-VM1:
	@echo "Construyendo Docker images para VM1..."
	sudo docker-compose -f $(COMPOSE_VM1) build

docker-VM1:
	@echo "Levantando contenedores en VM1 (Riploy / BD1 / Grupo2)..."
	docker-compose -f $(COMPOSE_VM1) up 

down-VM1:
	@echo "Deteniendo contenedores en VM1..."
	docker-compose -f $(COMPOSE_VM1) down

# ========== VM2 ==========
build-VM2:
	@echo "Construyendo Docker images para VM2..."
	sudo docker-compose -f $(COMPOSE_VM2) build

docker-VM2:
	@echo "Levantando contenedores en VM2 (Riploy / BD1 / Grupo2)..."
	docker-compose -f $(COMPOSE_VM2) up 

down-VM2:
	@echo "Deteniendo contenedores en VM2..."
	docker-compose -f $(COMPOSE_VM2) down

# ========== VM3 ==========
build-VM3:
	@echo "Construyendo Docker images para VM3..."
	sudo docker-compose -f $(COMPOSE_VM3) build

docker-VM3:
	@echo "Levantando contenedores en VM3 (Riploy / BD1 / Grupo2)..."
	docker-compose -f $(COMPOSE_VM3) up 

down-VM3:
	@echo "Deteniendo contenedores en VM3..."
	docker-compose -f $(COMPOSE_VM3) down

# ========== VM4 ==========
build-VM4:
	@echo "Construyendo Docker images para VM4..."
	sudo docker-compose -f $(COMPOSE_VM4) build

docker-VM4:
	@echo "Levantando contenedores en VM4 (Riploy / BD1 / Grupo2)..."
	docker-compose -f $(COMPOSE_VM4) up 

down-VM4:
	@echo "Deteniendo contenedores en VM4..."
	docker-compose -f $(COMPOSE_VM4) down

# ===================================================================

	
