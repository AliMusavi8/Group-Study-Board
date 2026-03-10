FROM node:24.13.1 AS build

WORKDIR /app
COPY package.json ./
RUN npm install

COPY . .
RUN npm run build

FROM nginx:1.29.5
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist/group-study-board /usr/share/nginx/html

EXPOSE 4200
