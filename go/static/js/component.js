// Frontend arch problems:
//   Components are irregular (no get/set/delete/update)
//   Solution: limit DRY and abstraction layers

const component = {
    renderer: null,
    message: null,
    media: null,
    init: async function (renderer, message, media) {
        this.renderer = renderer
        this.message = message.init()
        // Add user message send capability to chat input
        this.assignHandleChatInput()
        this.message.assignCallbacks(
            this.handleOpen.bind(this),
            this.handleError.bind(this),
            this.handleMessage.bind(this),
            this.handleClose.bind(this)
        )
        this.media = await media.init()
        // show local video
        this.renderer.videoAreaOverlay.addVideo(this.media.stream)
        // Add tracks to peer connection
        this.media.addTrack()
        // Tell peerConnection what to do when it recieves a candidate and track
        this.media.assignCallbacks(
            this.handleOnTrack.bind(this),
            this.handleIceCandidate.bind(this)
        )
        // Order media connecitons from backend
        this.orderMedia()
    },

    assignHandleChatInput: function () {
        const chatInputElement = this.renderer.chatInput.getElement()
        if (chatInputElement) {
            chatInputElement.onsubmit = (e) => {
                e.preventDefault()
                const formData = new FormData(e.target)
                const { message } = Object.fromEntries(formData)
                const msg = {
                    event: 'user-message',
                    data: message,
                }
                this.message.sendMessage(msg)
                e.target.reset()
                // this.renderer.chatLog.addMessage(message)
            }
        }
    },

    handleOnTrack: function (e) {
        if (e.track.kind === 'audio') {
            return
        }
        // Add a video to the screen for every track
        const videoElement = this.renderer.videoAreaOverlay.addVideo(
            e.streams[0]
        )
        // Videos are muted by default
        e.track.onmute = function () {
            videoElement.play()
        }
        // define this in the renderer
        e.streams[0].onremovetrack = () => {
            if (videoElement.parentNode) {
                videoElement.parentNode.removeChild(videoElement)
            }
        }
    },
    handleIceCandidate: function (e) {
        if (!e.candidate) {
            return
        }
        const message = {
            event: 'candidate',
            data: JSON.stringify(e.candidate),
        }
        this.message.sendMessage(message)
    },
    handleMessage: async function (event) {
        const message = JSON.parse(event.data)
        console.log({ received_message: message })
        if (!message) {
            return console.log('failed to parse message ', event.data)
        }
        if (message.event == 'candidate') {
            let candidate = JSON.parse(message.data)
            if (!candidate) {
                return console.log('failed to parse candidate')
            }
            this.media.addCandidate(candidate)
            return
        }

        if (message.event == 'offer') {
            let offer = JSON.parse(message.data)
            if (!offer) {
                return console.log('failed to parse answer')
            }
            this.media.setRemoteDescription(offer)
            // const answer = this.media.createAnswer()
            const answer = await this.media.createAnswer()
            this.media.setLocalDescription(answer)
            const msg = {
                event: 'answer',
                data: JSON.stringify(answer),
            }
            this.message.sendMessage(msg)
        }

        if (message.event == 'joined-room') {
            const chatLog = message.data.chat_log
            if (chatLog && chatLog.length) {
                let i = 0
                while (i < chatLog.length) {
                    this.renderer.chatLog.addMessage(chatLog[i])
                    i = i + 1
                }
            }
            return
        }

        if (message.event == 'created-room') {
            const urlParams = new URLSearchParams(window.location.search)
            urlParams.set('room', message.data)
            window.location.search = urlParams
            return
        }

        if (message.event == 'user-message') {
            this.renderer.chatLog.addMessage(message.data)
            return
        }
    },

    handleError: function (event) {
        this.renderer.chatLog.addMessage('ADMIN (to you): error: ' + event.data)
    },

    handleOpen: function () {
        this.renderer.chatLog.addMessage(
            'ADMIN (to you): able to receive messages'
        )
    },

    handleClose: function () {
        this.renderer.chatLog.addMessage('ADMIN (to you): connection closed')
    },

    orderMedia: function () {
        console.log('orderMessag')
        const message = {
            event: 'media_request',
            data: this.media.constraints,
        }
        this.message.sendMessage(message)
    },
}

export default component
