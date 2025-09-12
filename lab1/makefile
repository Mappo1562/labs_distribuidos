# Makefile - raíz del repo

CLIENTS := Franklin Lester Michael Trevor

PROTOS := Franklin/proto/franklin.proto \
          Lester/proto/lester.proto \
          Michael/proto/michael.proto \
          Trevor/proto/trevor.proto

.PHONY: all help protos modules build clean

all: protos modules build

help:
	@echo "Usos:"
	@echo "  make all      - genera protos, init/tidy de modulos, y compila clientes"
	@echo "  make protos   - genera los .pb.go desde los .proto"
	@echo "  make modules  - go mod init (si falta) y go mod tidy en cada cliente"
	@echo "  make build    - compila los clientes y coloca binarios en ./bin"
	@echo "  make clean    - borra bin/ y .pb.go generados"

# Genera go code a partir de cada proto encontrado
protos:
	@echo "Generando código Go desde archivos .proto..."
	@if [ -z "$(PROTOS)" ]; then \
		echo "  (no se encontraron archivos .proto)"; \
	else \
		for p in $(PROTOS); do \
			dir=$$(dirname $$p); \
			base=$$(basename $$p); \
			echo "  procesando $$p ..."; \
			( cd $$dir && protoc --go_out=. --go-grpc_out=. $$base ) || { echo "ERROR: protoc falló en $$p"; exit 1; }; \
		done; \
		echo "Generación de protos completada."; \
	fi

# Inicializa módulo si no existe y ejecuta go mod tidy en cada cliente
modules:
	@echo "Inicializando módulos y ejecutando 'go mod tidy' donde corresponda..."
	@for d in $(CLIENTS); do \
		if [ -d "$$d" ]; then \
			if [ ! -f "$$d/go.mod" ]; then \
				echo "  go mod init $$d (en $$d)"; \
				( cd $$d && go mod init $$d ) || { echo "ERROR: go mod init falló en $$d"; exit 1; }; \
			fi; \
			echo "  go mod tidy (en $$d)"; \
			( cd $$d && go mod tidy ) || { echo "ERROR: go mod tidy falló en $$d"; exit 1; }; \
		else \
			echo "  (no existe carpeta $$d, se omite)"; \
		fi; \
	done
	@echo "Módulos listos."

# Compila cada cliente que tenga un archivo cliente*.go
build:
	@echo "Compilando clientes..."
	@mkdir -p bin
	@for d in $(CLIENTS); do \
		if [ -d "$$d" ]; then \
			if ls $$d/cliente*.go >/dev/null 2>&1; then \
				echo "  compilando $$d -> bin/$$d"; \
				( cd $$d && go build -o ../bin/$$d . ) || { echo "ERROR: go build falló en $$d"; exit 1; }; \
			else \
				echo "  (no hay cliente*.go en $$d, se omite)"; \
			fi; \
		fi; \
	done
	@echo "Compilación finalizada. Binaries en ./bin/"

# Borra bin y los pb.go generados dentro de */proto/grpc-server/proto
clean:
	@echo "Limpiando bin y pb.go generados..."
	@rm -rf bin
	@for p in $(PROTOS); do \
		dir=$$(dirname $$p); \
		outdir=$${dir}/grpc-server/proto; \
		if [ -d "$$outdir" ]; then \
			echo "  borrando $$outdir/*.pb.go"; \
			rm -f $$outdir/*.pb.go || true; \
		fi; \
	done
	@echo "Limpieza completada."