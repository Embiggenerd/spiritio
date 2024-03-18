const component = {
    renderer: null,
    message: null,
    media: null,

    init: async function (render, message, media) {
        this.renderer = render()
        this.message = message.init()
        // Add user message send capability to chat input
        this.assignHandleChatInput()
        this.message.assignCallbacks(
            this.handleOpen.bind(this),
            this.handleMessageError.bind(this),
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
        const chatFormElement = this.renderer.chatInput.getElement()
        if (chatFormElement) {
            chatFormElement.onsubmit = this.onChatSubmit.bind(this)
            const chatInputElement = this.renderer.chatInput.getInputElement()
            this.setOnChatKeyDown(chatInputElement)
        }
    },

    setOnChatKeyDown: function (chatInputElement) {
        if (!chatInputElement) {
            throw new Error('chat input element not found')
        }

        const up = 38
        const down = 40
        const commandLog = this.getCommandLog()
        const lenCommands = commandLog.length - 1
        const empty = ''
        let i = -1

        chatInputElement.onkeydown = (event) => {
            if (event.keyCode === up || event.keyCode === down) {
                if (event.keyCode === up) {
                    i--
                }
                if (event.keyCode === down) {
                    i++
                }
                if (i < -1) {
                    i = lenCommands
                }
                if (i > lenCommands) {
                    i = -1
                }
                if (i === -1) {
                    event.target.value = empty
                } else {
                    event.target.value = commandLog[i]
                }
            }
        }
    },

    onChatSubmit: function (event) {
        try {
            event.preventDefault()
            const formData = new FormData(event.target)
            const { message } = Object.fromEntries(formData)
            if (message.startsWith('/')) {
                const commandConfig = this.parseUserCommand(message)

                const work = {
                    order: '',
                    details: {},
                }

                work.order = commandConfig.workOrder
                commandConfig.args.forEach((a) => {
                    work.details[a.name] = a.value
                })
                this.addToCommandLog(message)
                this.setOnChatKeyDown(this.renderer.chatInput.getInputElement())
                this.orderWork(work.order, work.details)
            } else {
                this.orderWork('user_message', message)
            }
            event.target.reset()
        } catch (e) {
            this.handleError(e)
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
            throw new Error('failed to parse message ' + event.data)
        }
        console.log('recieved message', message.type, message.data)
        if (message.type == 'event') {
            await this.handleEvent(message.data.event, message.data.data)
        }
        if (message.type == 'question') {
            this.handleQuestion(message.data.ask)
        }
    },

    handleMessageError: function (event) {
        this.handeError(event)
    },

    handleError: function (error) {
        console.error(error)
        const message = {
            text: error,
            from: 'ADMIN (to you)',
        }
        if (error instanceof Event) {
            message.text = error.data
        }
        if (error instanceof Error) {
            message.text = error.message
        }
        if (error.text) {
            message.text = error.text
        }
        this.renderer.chatLog.addMessage(message)
    },

    handleOpen: function () {
        this.renderer.chatLog.addMessage({
            text: 'able to receive messages',
            from: 'ADMIN (to you)',
        })
    },

    handleClose: function () {
        // this.orderWork('close_connection', '')
        this.media.closePeerConnection()
        this.renderer.chatLog.addMessage({
            text: 'unable to receive messages',
            from: 'ADMIN (to you)',
        })
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

    handleEvent: async function (event, data) {
        if (event == 'candidate') {
            let candidate = JSON.parse(data)
            if (!candidate) {
                throw new Error('failed to parse candidate')
            }
            this.media.addCandidate(candidate)
            return
        }

        if (event == 'offer') {
            let offer = JSON.parse(data)
            if (!offer) {
                throw new Error('failed to parse answer')
            }
            this.media.setRemoteDescription(offer)
            const answer = await this.media.createAnswer()
            this.media.setLocalDescription(answer)

            this.orderWork('answer', JSON.stringify(answer))
        }

        if (event === 'joined_room') {
            const chatLog = data.chat_log
            if (chatLog && chatLog.length) {
                let i = 0
                while (i < chatLog.length) {
                    this.renderer.chatLog.addMessage(chatLog[i])
                    i = i + 1
                }
            }
            return
        }

        if (event === 'created_room') {
            const urlParams = new URLSearchParams(window.location.search)
            urlParams.set('room', data)
            window.location.search = urlParams
            return
        }

        if (event === 'user_message') {
            this.renderer.chatLog.addMessage(data)
            return
        }

        if (event === 'user_logged_in') {
            localStorage.setItem('access_token', data.access_token)
            const message = {
                from: `ADMIN (to you)`,
                text: `logged in as ${data.name}`,
            }
            this.renderer.chatLog.addMessage(message)
        }

        if (event === 'user_name_change') {
            this.renderer.chatLog.addMessage({
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
    },

    handleQuestion: function (ask) {
        if (ask === 'access_token') {
            const accessToken = localStorage.getItem('access_token')
            this.orderWork('validate_access_token', accessToken || '')
            return
        }

        if (ask === 'credentials') {
            const text = `We have created a new, unverified account for you. If you would like to add a password, type "/set password <password>", and you can change your name. If you would like to log into a different account, type "/login <name> <password>".`
            const message = {
                text,
                from: 'ADMIN (to you)',
            }
            this.renderer.chatLog.addMessage(message)
            return
        }
    },

    // Parses a user message starting with '/' into a work order
    parseUserCommand: function (command) {
        let workOrderKey = ''
        const commandConfig = {
            'set password': {
                workOrder: 'set_user_password',
                args: [
                    {
                        name: 'password',
                        regex: /[A-Za-z0-9_@./#&+!$_*+-]/,
                        value: '',
                    },
                ],
            },
            'set name': {
                workOrder: 'set_user_name',
                args: [
                    {
                        name: 'name',
                        regex: /[A-Za-z0-9_@./#&+!$_*+-]/,
                        value: '',
                    },
                ],
            },
            login: {
                workOrder: 'validate_user_name_password',
                args: [
                    // order of these matters
                    {
                        name: 'name',
                        regex: /[A-Za-z0-9_@./#&+!$_*+-]/,
                        value: '',
                    },
                    {
                        name: 'password',
                        regex: /[A-Za-z0-9_@./#&+!$_*+-]/,
                        value: '',
                    },
                ],
            },
        }
        const commandChar = '/'
        const allLettersRegex = /[a-zA-Z]/
        let i = 0

        parse()
        let argsCount = 0
        let argsRequired = 0

        commandConfig[workOrderKey].args.forEach((a) => {
            if (a.value) {
                argsCount++
            }
            argsRequired++
        })
        if (argsCount !== argsRequired) {
            throw new Error(
                `wrong argument number: have ${argsCount}, want ${argsRequired}`
            )
        }
        return commandConfig[workOrderKey]

        function parse() {
            if (match(commandChar)) {
                eat(commandChar)
            }
            parseCommand(readWhileMatching(allLettersRegex))
            skipWhitespace()
            parseArguments()
        }
        // Find any words that are in the list of possible commands
        function parseCommand(word) {
            // if (commandElements.includes(word.toLowerCase())) {
            workOrderKey = workOrderKey + ' ' + word
            workOrderKey = workOrderKey.trim()
            if (
                Object.keys(commandConfig).includes(
                    workOrderKey.toLowerCase()
                ) ||
                i === command.length
            ) {
                return
            }
            skipWhitespace()
            parseCommand(readWhileMatching(allLettersRegex))
        }
        // Figure out the arguments according to config
        function parseArguments() {
            if (!commandConfig.hasOwnProperty(workOrderKey)) {
                throw new Error('no such command')
            }
            const args = commandConfig[workOrderKey].args
            let j = 0
            if (args) {
                while (j < args.length) {
                    skipWhitespace()
                    args[j].value = readWhileMatching(args[j].regex)
                    j++
                }
            }
            skipWhitespace()
            const next = i + 1
            if (command[next]) {
                throw new Error('Too many arguments')
            }
        }

        function eat(str) {
            if (match(str)) {
                i += str.length
            } else {
                throw new Error(`Parse error: expecting "${str}"`)
            }
        }

        function match(str) {
            return command.slice(i, i + str.length) === str
        }

        function readWhileMatching(regex) {
            let startIndex = i
            while (i < command.length && regex.test(command[i])) {
                i++
            }
            return command.slice(startIndex, i)
        }

        function skipWhitespace() {
            readWhileMatching(/[\s\n]/)
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
