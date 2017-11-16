#include <grpc++/grpc++.h>

#include "ping.grpc.pb.h"
#include "ping.pb.h"

using namespace pinger;

class PingerImpl final : public Pinger::Service {
 public:
  explicit PingerImpl() {
  }

  grpc::Status Ping(grpc::ServerContext* context, const PingRequest* request, PingResponse* response) {
    response->set_payload("k19{D~3+ld");
    return grpc::Status::OK;
  }

  grpc::Status PingStream(grpc::ServerContext* context, grpc::ServerReaderWriter< PingResponse, PingRequest>* stream) {
    PingRequest req;
    PingResponse resp;
    resp.set_payload("k19{D~3+ld");
    while (stream->Read(&req)) {
      stream->Write(resp);
    }
    return grpc::Status::OK;
  }
};

void RunServer() {
  std::string server_address("0.0.0.0:50051");
  PingerImpl service;

  grpc::ServerBuilder builder;
  builder.AddListeningPort(server_address, grpc::InsecureServerCredentials());
  builder.RegisterService(&service);
  std::unique_ptr<grpc::Server> server(builder.BuildAndStart());
  std::cout << "Server listening on " << server_address << std::endl;
  server->Wait();
}

int main() {
  RunServer();
  return 0;
}
