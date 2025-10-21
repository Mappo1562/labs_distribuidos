# Archivos docker-compose por VM
COMPOSE_VM1=Dist042/docker-compose.yml
COMPOSE_VM2=Dist043/docker-compose.yml
COMPOSE_VM3=Dist044/docker-compose.yml
COMPOSE_VM4=Dist101/docker-compose.yml

create-network:
	@docker network inspect grpc-net >/dev/null 2>&1 || \
	echo "🌐 Creando red grpc-net..." && docker network create grpc-net

# ========== VM1 ==========
build-VM1:
	@echo "Construyendo Docker images para VM1..."
	sudo docker-compose -f $(COMPOSE_VM1) build

docker-VM1: create-network
	@echo "Levantando contenedores en VM1 (Riploy / BD1 / Grupo2)..."
	docker-compose -f $(COMPOSE_VM1) up -d

down-VM1:
	@echo "Deteniendo contenedores en VM1..."
	docker-compose -f $(COMPOSE_VM1) down

# ========== VM2 ==========
build-VM2:
	@echo "Construyendo Docker images para VM2..."
	sudo docker-compose -f $(COMPOSE_VM2) build

docker-VM2: create-network
	@echo "Levantando contenedores en VM2 (Riploy / BD1 / Grupo2)..."
	docker-compose -f $(COMPOSE_VM2) up -d

down-VM2:
	@echo "Deteniendo contenedores en VM2..."
	docker-compose -f $(COMPOSE_VM2) down

# ========== VM3 ==========
build-VM3:
	@echo "Construyendo Docker images para VM3..."
	sudo docker-compose -f $(COMPOSE_VM3) build

docker-VM3: create-network
	@echo "Levantando contenedores en VM3 (Riploy / BD1 / Grupo2)..."
	docker-compose -f $(COMPOSE_VM3) up -d

down-VM3:
	@echo "Deteniendo contenedores en VM3..."
	docker-compose -f $(COMPOSE_VM3) down

# ========== VM4 ==========
build-VM4:
	@echo "Construyendo Docker images para VM4..."
	sudo docker-compose -f $(COMPOSE_VM4) build

docker-VM4: create-network
	@echo "Levantando contenedores en VM4 (Riploy / BD1 / Grupo2)..."
	docker-compose -f $(COMPOSE_VM4) up -d

down-VM4:
	@echo "Deteniendo contenedores en VM4..."
	docker-compose -f $(COMPOSE_VM4) down

# ===================================================================


# ============================
# Atajos de Prueba Rápida
# ============================

build-all:
	@echo "🧱 Construyendo todas las imágenes Docker..."
	docker-compose -f $(COMPOSE_VM1) build
	docker-compose -f $(COMPOSE_VM2) build
	docker-compose -f $(COMPOSE_VM3) build
	docker-compose -f $(COMPOSE_VM4) build

up-all: docker-VM4 docker-VM1 docker-VM2 docker-VM3 
	@echo "🌐 Todos los contenedores distribuidos están arriba ✅"
	