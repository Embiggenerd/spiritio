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
        // Order media capabilities from backend
        this.orderMedia()
    },

    assignHandleChatInput: function () {
        const chatInputElement = this.renderer.chatInput.getElement()
        if (chatInputElement) {
            chatInputElement.onsubmit = (e) => {
                e.preventDefault()
                const formData = new FormData(e.target)
                const { message } = Object.fromEntries(formData)

                this.orderWork('user_message', message)
                e.target.reset()
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
        this.orderWork('candidate', JSON.stringify(e.candidate))
    },
    handleMessage: async function (event) {
        const message = JSON.parse(event.data)
        if (!message) {
            return console.log('failed to parse message ', event.data)
        }
        if (message.type == 'event') {
            await this.handleEvent(message.data)
        }
        if (message.type == 'question') {
            this.handleQuestoin(message.data)
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
        this.orderWork('close_connection', '')
        this.renderer.chatLog.addMessage(
            'ADMIN (to you): unable to receive messages'
        )
    },

    orderMedia: function () {
        this.orderWork('media_request', this.media.constraints)
    },

    orderWork: function (order, details) {
        const message = {
            order,
            details,
        }
        this.message.sendMessage(message)
    },

    handleEvent: async function (event) {
        if (event.event == 'candidate') {
            let candidate = JSON.parse(event.data)
            if (!candidate) {
                return console.log('failed to parse candidate')
            }
            this.media.addCandidate(candidate)
            return
        }

        if (event.event == 'offer') {
            let offer = JSON.parse(event.data)
            if (!offer) {
                return console.log('failed to parse answer')
            }
            this.media.setRemoteDescription(offer)
            // const answer = this.media.createAnswer()
            const answer = await this.media.createAnswer()
            this.media.setLocalDescription(answer)

            this.orderWork('answer', JSON.stringify(answer))
        }

        if (event.event == 'joined_room') {
            const chatLog = event.data.chat_log
            if (chatLog && chatLog.length) {
                let i = 0
                while (i < chatLog.length) {
                    this.renderer.chatLog.addMessage(chatLog[i])
                    i = i + 1
                }
            }
            return
        }

        if (event.event == 'created_room') {
            const urlParams = new URLSearchParams(window.location.search)
            urlParams.set('room', event.data)
            window.location.search = urlParams
            return
        }

        if (event.event == 'user_message') {
            this.renderer.chatLog.addMessage(event.data)
            return
        }
    },
    handleQuestoin: function (question) {
        console.log({ question })
    },
}

export default component
