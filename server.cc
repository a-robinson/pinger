#include <grpc++/grpc++.h>

#include "grpc.h"
#include "ping.grpc.pb.h"
#include "ping.pb.h"

using namespace pinger;

Client NewClient() {
  std::unique_ptr<Pinger::Stub> stub = Pinger::NewStub(grpc::CreateChannel(
              "localhost:50051", grpc::InsecureChannelCredentials()));
  return (void*)(stub.release());
}

void ClientFree(Client cStub) {
  Pinger::Stub* stub = (Pinger::Stub*)cStub;
  delete stub;
}

int ClientSend(Client cStub) {
  Pinger::Stub* stub = (Pinger::Stub*)cStub;
  grpc::ClientContext context;
  PingRequest request;
  request.set_payload("k19{D~3+ld");
  PingResponse response;
  grpc::Status status = stub->Ping(&context, request, &response);
  if (!status.ok()) {
    abort();
  }
  return response.payload().length();
}

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
