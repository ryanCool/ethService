.PHONY: run clean

run:
	@echo "Making devenv..."
	cd devenv ; docker-compose pull ; docker-compose up -d --build; cd ..

clean:
	@echo "Cleaning devenv..."
	cd devenv ; docker-compose down ; cd ..