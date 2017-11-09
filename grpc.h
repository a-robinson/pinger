#ifndef GRPC_H
#define GRPC_H

#ifdef __cplusplus
extern "C" {
#endif

typedef void* Client;
Client NewClient(void);
void ClientFree(Client);
int ClientSend(Client);

#ifdef __cplusplus
}
#endif

#endif
