FROM golang:1.15

#Install cmake here for libson
RUN apt-get update && apt-get -y install cmake

#Install openssl for libson
RUN apt-get install libssl-dev

#Install git to clone mongocrypt repo
RUN apt-get -y install git

#Step1: mongocrypt instalation -> Install libson here
RUN git clone https://github.com/mongodb/mongo-c-driver &&\
    cd mongo-c-driver && mkdir cmake-build && cd cmake-build  \ -DCMAKE_INSTALL_RPATH_USE_LINK_PATH=yes  &&\
    cmake -DENABLE_MONGOC=OFF -DCMAKE_INSTALL_PREFIX="/usr/local" ../ &&\
    cmake -DENABLE_SHARED_BSON=ON .. &&\
    cmake --build . --target install

#Step2:
RUN cd /usr/local/ && git clone https://github.com/mongodb/libmongocrypt.git &&\
   cd /usr/local/libmongocrypt &&\
   mkdir cmake-build && cd cmake-build \ -DCMAKE_INSTALL_RPATH_USE_LINK_PATH=yes &&\
    cmake -DENABLE_SHARED_BSON=ON .. &&\
    cmake --build . --target install

RUN cd mongo-c-driver/cmake-build \ -DCMAKE_INSTALL_RPATH_USE_LINK_PATH=yes && cmake -DENABLE_AUTOMATIC_INIT_AND_CLEANUP=OFF -DENABLE_MONGOC=ON -DENABLE_CLIENT_SIDE_ENCRYPTION=ON .. && cmake --build . --target install

WORKDIR /go/src/chrome-extension-back-end

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -tags cse /go/src/chrome-extension-back-end/cmd/api/main.go

EXPOSE 8081

CMD ["./main"]
