FROM golang:1.20

# RUN apt-get update -qq && apt-get install -y nodejs npm

# RUN npm install -g nodemon

EXPOSE 3000

RUN mkdir twitter

WORKDIR /twitter

COPY . /twitter

RUN go get -u

CMD ["go", "run", "."]