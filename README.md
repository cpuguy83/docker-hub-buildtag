DockerHub buildtag  [![Docker Build Status](http://72.14.176.28/cpuguy83/docker-hub-buildtag)](https://registry.hub.docker.com/u/cpuguy83/docker-hub-buildtag)

Generates build status tags for Dockerhub automated builds as seen above

Usage
```
docker run -d --name redis redis
docker run --link redis:db -d cpuguy83/docker-hub-buildtag -redis db:6379
```

To embed the generated badge, see the raw output of this README.md
