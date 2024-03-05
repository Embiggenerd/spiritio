const message = {
    path: '/ws',
    scheme: window.location.protocol == 'https:' ? 'wss' : 'ws',
    conn: null,
    init: function () {
        this.conn = this.connect()
        return this
    },
    sendMessage: function (message) {
        console.log({ sending_message: message })
        this.conn.send(JSON.stringify(message))
    },
    connect: function () {
        const url =
            this.scheme +
            '://' +
            window.location.host +
            this.path +
            window.location.search
        return new WebSocket(url)
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

export default message
