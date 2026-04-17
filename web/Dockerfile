# syntax=docker/dockerfile:1.7

# ---- Build stage --------------------------------------------------------
FROM node:20-alpine AS build

WORKDIR /src

# Install deps first for better layer caching.
COPY package.json package-lock.json* ./
RUN npm ci

COPY . .
RUN npm run build

# ---- Runtime stage ------------------------------------------------------
FROM nginx:1.27-alpine

COPY --from=build /src/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
