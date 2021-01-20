FROM golang:1.15

#Install cmake here for libson
RUN apt-get update && apt-get -y install cmake

#Install openssl for libson
RUN apt-get install libssl-dev

#Install git to clone mongocrypt repo
RUN apt-get -y install git

#Step1: mongocrypt instalation -> Install libson here
RUN git clone https://github.com/mongodb/mongo-c-driver &&\
    cd mongo-c-driver && mkdir cmake-build && cd cmake-build  &&\
    cmake  -DENABLE_MONGOC=OFF -DCMAKE_PREFIX_PATH="/usr/local" ../ &&\
    cmake --build . --target install

#Step2: install lib mongo crypt
RUN git clone https://github.com/mongodb/libmongocrypt.git &&\
   cd libmongocrypt &&\
   mkdir cmake-build && cd cmake-build &&\
    cmake  -DENABLE_SHARED_BSON=ON -DCMAKE_PREFIX_PATH="/usr/local" ../  &&\
    cmake --build . --target install

#Step 3: enable client-field encryption
RUN cd mongo-c-driver/cmake-build  &&\
    cmake -DENABLE_AUTOMATIC_INIT_AND_CLEANUP=OFF -DENABLE_MONGOC=ON -DENABLE_CLIENT_SIDE_ENCRYPTION=ON .. &&\
     cmake --build . --target install

#Step 4: copy installed libraries from /usr/local/lin to /usr/lib expected directory;
RUN cd /usr/local/lib && cp libmongocrypt.so.0 libbson-1.0.so.0 /usr/lib

WORKDIR /go/src/chrome-extension-back-end

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -tags cse /go/src/chrome-extension-back-end/cmd/api/main.go

EXPOSE 8081

CMD ["./main"]
