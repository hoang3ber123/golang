To run project:
1: Create network: kong-network
_ cmd: docker network create kong-network
2: Build Kong container:
_ cmd: docker-compose build
3: Run Kong container:
_ cmd: docker-compose up
4: Build Auth-services container:
_ cmd: docker-compose build
5: Run Auth-services container:
_ cmd: docker-compose up
