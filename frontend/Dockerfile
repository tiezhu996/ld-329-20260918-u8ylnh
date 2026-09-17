FROM node:20 AS builder
WORKDIR /app
COPY package*.json ./
RUN npm config set registry https://registry.npmmirror.com && if [ -f package-lock.json ]; then npm ci --no-audit --omit=optional; else npm install --no-audit --omit=optional; fi
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
