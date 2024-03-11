# base frontend #######################
FROM node:18-alpine AS frontend-base

WORKDIR /app

# RUN apk add --no-cache libc6-compat
COPY package.json package-lock.json ./
RUN npm install


# build frontend #######################
FROM node:18-alpine AS frontend-builder

WORKDIR /app

COPY --from=frontend-base /app/node_modules ./node_modules
COPY . .
RUN npm run build


# build server #######################
FROM golang:1.19.13-bullseye as server-builder

WORKDIR /app

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/gomod-cache \
    go mod download

COPY . .
COPY --from=frontend-builder /app/dist ./dist

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    go build -trimpath -ldflags="-w -s" -o bin/server .

# run server #######################
FROM golang:1.19.13-bullseye as runner

WORKDIR /app

COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.7.2 /lambda-adapter /opt/extensions/lambda-adapter
COPY --from=server-builder /app/bin/server .
EXPOSE 8080
ENV PORT 8080

CMD ["/app/server"]
