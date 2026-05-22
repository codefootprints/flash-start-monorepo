# Stage 1: Build
FROM node:25-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci 
COPY . .
ENV NODE_OPTIONS="--max-old-space-size=4096"
RUN npm run build

# Stage 2: Production (Ganti serve dengan Nginx)
FROM nginx:alpine
# Copy hasil build ke folder default Nginx
COPY --from=build /app/dist /usr/share/nginx/html
# Tambahkan konfigurasi agar client-side routing React (React Router) tidak 404
RUN echo "server { listen 80; location / { root /usr/share/nginx/html; index index.html; try_files \$uri \$uri/ /index.html; } }" > /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]