# ---------- Build ----------
FROM node:22-alpine AS builder

# git is required by some npm packages and Nuxt prepare hook.
RUN apk add --no-cache git

WORKDIR /app

# Copy package manifest (lockfile is optional — generated on the fly if absent).
COPY package*.json ./

# `--ignore-scripts` skips Nuxt's postinstall (which needs source files).
# We run `nuxt prepare` manually right after copying the source.
RUN npm install --no-audit --no-fund --legacy-peer-deps --ignore-scripts

# Copy source.
COPY . .

# Generate Nuxt types & build the SSR bundle.
ENV NODE_ENV=production
RUN npx nuxt prepare && npx nuxt build

# ---------- Runtime ----------
FROM node:22-alpine AS runner

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# Nuxt build produces a self-contained .output/ directory.
COPY --from=builder /app/.output ./.output
COPY --from=builder /app/package.json ./package.json

ENV NODE_ENV=production
ENV NITRO_PORT=3000
ENV NITRO_HOST=0.0.0.0

USER app
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
