/**
 * @type {import("../../types").MessageService}
 */
const messageService = {
    webSocket: null,
    path: '/ws',
    scheme: window.location.protocol == 'https:' ? 'wss' : 'ws',
    conn: null,
    init: function (wsClass) {
        this.webSocket = wsClass || WebSocket
        this.conn = this.connect()
        return this
    },
    sendMessage: function (message) {
        this.conn.send(JSON.stringify(message))
    },
    connect: function () {
        const url =
            this.scheme +
            '://' +
            window.location.host +
            this.path +
            window.location.search
        return new this.webSocket(url)
    },
    assignCallbacks: function (
        handleOpen,
        handleError,
        handleMessage,
        handleClose
    ) {
        this.conn.onopen = handleOpen
        this.conn.onerror = handleError
        this.conn.onmessage = handleMessage
        this.conn.onclose = handleClose
    },
}

export default messageService
