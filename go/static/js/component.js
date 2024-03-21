import parser from './utils/parser.js'

/**
 * @type {import("../types").Component}
 */
const component = {
    renderer: null,
    mediaService: null,
    messageService: null,
    async init(render, messageService, mediaService) {
        try {
            this.mediaService = mediaService
            this.renderer = render()
            this.messageService = messageService.init(WebSocket)
            // Add user message send capability to chat input
            this.assignHandleChatInput()
            this.messageService.assignCallbacks(
                this.handleOpen.bind(this),
                this.handleMessageError.bind(this),
                this.handleMessage.bind(this),
                this.handleClose.bind(this)
            )
        } catch (e) {
            this.handleError(e)
        }
    },

    assignHandleChatInput() {
        try {
            const chatFormElement = this.renderer?.chatInput.getElement()
            if (chatFormElement) {
                chatFormElement.onsubmit = this.onChatSubmit.bind(this)
                const chatInputElement =
                    this.renderer?.chatInput.getInputElement()
                this.setOnChatKeyDown(chatInputElement)
            }
        } catch (e) {
            this.handleError(e)
        }
    },

    setOnChatKeyDown: function (chatInputElement) {
        try {
            if (!chatInputElement) {
                throw new Error('chat input element not found')
            }

            const up = 'ArrowUp'
            const down = 'ArrowDown'
            const commandLog = this.getCommandLog()
            const lenCommands = commandLog.length - 1
            const empty = ''
            let i = -1

            chatInputElement.onkeydown = (event) => {
                const target = event.target
                const assertTargetInputElement =
                    /** @type {HTMLInputElement | null }  */ (target)

                if (event.key === up || event.key === down) {
                    if (event.key === up) {
                        i--
                    }
                    if (event.key === down) {
                        i++
                    }
                    if (i < -1) {
                        i = lenCommands
                    }
                    if (i > lenCommands) {
                        i = -1
                    }
                    if (i === -1 && target) {
                        if (assertTargetInputElement)
                            assertTargetInputElement.value = empty
                    } else {
                        if (assertTargetInputElement)
                            assertTargetInputElement.value = commandLog[i]
                    }
                }
            }
        } catch (e) {
            this.handleError(e)
        }
    },

    onChatSubmit: function (event) {
        try {
            event.preventDefault()
            let formData
            if (event.currentTarget)
                formData = new FormData(event.currentTarget)

            let message
            if (formData) {
                message = Object.fromEntries(formData).message
            }
            if (typeof message === 'string' && message.startsWith('/')) {
                let commandConfig
                commandConfig = parser().parseUserCommand(message)

                /**
                 * @type {import("../types").WorkOrder}
                 */
                const work = {
                    order: '',
                    details: {},
                }

                work.order = commandConfig.workOrder
                commandConfig.args.forEach((a) => {
                    work.details[a.name] = a.value
                })
                this.addToCommandLog(message)
                this.setOnChatKeyDown(
                    this.renderer?.chatInput.getInputElement()
                )
                this.orderWork(work)
            } else {
                this.orderWork({ order: 'user_message', details: message })
            }
            event.target.reset()
        } catch (e) {
            this.handleError(e)
        }
    },

    handleOnTrack: function (event) {
        try {
            if (event.track.kind === 'audio') {
                return
            }
            const stream = event.streams[0]
            if (stream) {
                this.orderWork({
                    order: 'identify_streamid',
                    details: stream.id,
                })
            }
            // Add a video to the screen for every track
            const videoElement = this.renderer?.videoArea.addVideo(
                event.streams[0]
            )
            // Videos are muted by default
            event.track.onmute = function () {
                videoElement.play()
            }
            // Remove the video wrapper around the video to remove text overlay
            event.streams[0].onremovetrack = () => {
                if (videoElement.parentNode) {
                    videoElement.parentNode.remove()
                }
            }
        } catch (e) {
            this.handleError(e)
        }
    },

    handleIceCandidate: function (e) {
        if (!e.candidate) {
            return
        }
        this.orderWork({
            order: 'candidate',
            details: JSON.stringify(e.candidate),
        })
    },
    /**
     * Handles all message types
     * @param {any} event
     */
    handleMessage: async function (event) {
        try {
            const message = JSON.parse(event.data)
            if (!message) {
                throw new Error('failed to parse message ' + event.data)
            }
            if (message.type == 'event') {
                await this.handleEvent(message.data.event, message.data.data)
            }
            if (message.type == 'question') {
                this.handleQuestion(message.data.ask)
            }
        } catch (e) {
            this.handleError(e)
        }
    },

    handleMessageError: function (event) {
        this.handleError(event)
    },

    handleError: function (error) {
        console.error(error)
        const message = {
            text: error,
            from: 'ADMIN (to you)',
        }
        if (error.data) {
            message.text = error.data
        }
        if (error instanceof Error) {
            message.text = error.message
        }
        if (error.text) {
            message.text = error.text
        }
        this.renderer?.chatLog.addMessage(message)
    },

    handleOpen: function () {
        this.renderer?.chatLog.addMessage({
            text: 'able to receive messages',
            from: 'ADMIN (to you)',
        })
    },

    handleClose: function () {
        this.renderer?.videoArea.removeRemote()
        this.renderer?.chatLog.addMessage({
            text: 'unable to receive messages',
            from: 'ADMIN (to you)',
        })
    },

    orderMedia: function () {
        this.orderWork({
            order: 'media_request',
            details: this.mediaService?.constraints,
        })
    },

    orderWork: function (workOrder) {
        this.messageService?.sendMessage(workOrder)
    },

    handleEvent: async function (event, data) {
        try {
            // Server will send offers regardless if we ask
            if (this.mediaService?.permissionsGranted) {
                if (event == 'candidate') {
                    let candidate = JSON.parse(data)
                    if (!candidate) {
                        throw new Error('failed to parse candidate')
                    }
                    this.mediaService.addCandidate(candidate)
                    return
                }

                if (event == 'offer') {
                    let offer = JSON.parse(data)
                    if (!offer) {
                        throw new Error('failed to parse answer')
                    }
                    this.mediaService.setRemoteDescription(offer)
                    const answer = await this.mediaService.createAnswer()
                    this.mediaService.setLocalDescription(answer)

                    this.orderWork({
                        order: 'answer',
                        details: JSON.stringify(answer),
                    })
                }
            }
            if (event === 'joined_room') {
                const chatLog = data.chat_log
                if (chatLog && chatLog.length) {
                    let i = 0
                    while (i < chatLog.length) {
                        this.renderer?.chatLog.addMessage(chatLog[i])
                        i = i + 1
                    }
                }
                if (this.mediaService) {
                    this.mediaService = await this.mediaService.init()
                    if (this.mediaService?.permissionsGranted) {
                        // show local video
                        this.renderer?.videoArea.addVideo(
                            this.mediaService.stream
                        )
                        // Add tracks to peer connection
                        this.mediaService.addTrack()
                        // Tell peerConnection what to do when it recieves a candidate and track
                        this.mediaService.assignCallbacks(
                            this.handleOnTrack.bind(this),
                            this.handleIceCandidate.bind(this)
                        )
                        if (this.mediaService.permissionsGranted) {
                            this.orderMedia()
                        }
                    }
                    return
                }
            }

            if (event === 'created_room') {
                const urlParams = new URLSearchParams(window.location.search)
                urlParams.set('room', data)
                window.location.search = urlParams.toString()
                return
            }

            if (event === 'user_message') {
                this.renderer?.chatLog.addMessage(data)
                return
            }

            if (event === 'user_logged_in') {
                localStorage.setItem('access_token', data.access_token)
                const message = {
                    from: `ADMIN (to you)`,
                    text: `logged in as ${data.name}`,
                }
                this.renderer?.chatLog.addMessage(message)
            }

            if (event === 'user_name_change') {
                this.renderer?.chatLog.addMessage({
                    from: 'ADMIN (to you): ',
                    text: `Name changed to ${data}`,
                })
            }

            if (event === 'error') {
                console.error(data)
                if (data.public) {
                    this.handleError(data.message)
                }
            }

            if (event === 'streamid_user_name') {
                this.renderer?.videoArea.identifyStream(
                    data.stream_id,
                    data.name
                )
            }

            if (event === 'user_joined_chat') {
                const text = `${data} has joined the chat`
                const message = {
                    text,
                    from: 'ADMIN',
                }
                this.renderer?.chatLog.addMessage(message)
            }
        } catch (e) {
            this.handleError(e)
        }
    },

    handleQuestion: function (ask) {
        if (ask === 'access_token') {
            const accessToken = localStorage.getItem('access_token')
            this.orderWork({
                order: 'validate_access_token',
                details: accessToken || '',
            })
            return
        }

        if (ask === 'credentials') {
            const text = `We have created a new, unverified account for you. If you would like to add a password, type "/set password <password>", and you can change your name. If you would like to log into a different account, type "/login <name> <password>".`
            const message = {
                text,
                from: 'ADMIN (to you)',
            }
            this.renderer?.chatLog.addMessage(message)
            return
        }
    },

    getCommandLog: function () {
        return JSON.parse(window.localStorage.getItem('command_log') || '[]')
    },

    addToCommandLog: function (command) {
        const commandLog = this.getCommandLog()
        commandLog.push(command)
        window.localStorage.setItem('command_log', JSON.stringify(commandLog))
    },
}

export default component
