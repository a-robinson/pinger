#include <grpc++/grpc++.h>

#include "grpc.h"
#include "ping.grpc.pb.h"
#include "ping.pb.h"

using namespace main;

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
