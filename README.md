To run project:
# Create network
1: Create network: kong-network
_ cmd: docker network create kong-network

# Run api gateway
2: Build Kong container:
_ cmd: docker-compose build
3: Run Kong container:
_ cmd: docker-compose up

# Run auth-services
4: Build Auth-services container:
_ cmd: docker-compose build
5: Run Auth-services container:
_ cmd: docker-compose up
