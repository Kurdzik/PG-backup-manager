FROM node:20.18.0-bullseye

RUN apt update && apt-get clean
WORKDIR /frontend

COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile

COPY . .

ARG NEXT_PUBLIC_BACKEND_URL
ENV NEXT_PUBLIC_BACKEND_URL=${NEXT_PUBLIC_BACKEND_URL}

RUN yarn build

EXPOSE 3000

CMD ["yarn", "start"]