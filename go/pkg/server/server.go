package server

// import (
// 	"net/http"
// 	"time"

// 	"github.com/Embiggenerd/spiritio/pkg/config"
// 	"github.com/Embiggenerd/spiritio/pkg/logger"
// 	"github.com/Embiggenerd/spiritio/pkg/rooms"
// )

// type APIServer struct {
// 	server       *http.Server
// 	roomsService rooms.RoomsService
// }

// func NewServer(cfg *config.Config, log logger.Logger, rooms rooms.RoomsService) *APIServer {
// 	srvr := &http.Server{
// 		Addr:              cfg.Server.Addr,
// 		// ReadHeaderTimeout: 3 * time.Second,
// 		ReadHeaderTimeout cf.Server.ReadHeaderTimeout
// 	}

// 	return &APIServer{
// 		listenAddr: listenAddr,
// 		store:      store,
// 	}
// }
