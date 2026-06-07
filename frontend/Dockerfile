# ---------- Build ----------
FROM node:22-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm install --no-audit --no-fund

COPY . .
ENV NODE_ENV=production
RUN npm run build

# ---------- Run ----------
FROM node:22-alpine AS runner

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /app/.output ./.output
COPY --from=builder /app/package.json ./package.json

ENV NODE_ENV=production
ENV NITRO_PORT=3000
ENV NITRO_HOST=0.0.0.0

USER app
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
