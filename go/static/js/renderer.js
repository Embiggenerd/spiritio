const chatMessage = {
    type: 'div',
    classList: ['chat-message'],
    create: function (text, from) {
        const element = document.createElement(this.type)
        this.classList.forEach((c) => {
            element.classList.add(c)
        })

        if (message) {
            const p = document.createElement('p')
            p.innerText = `${from}: ${text}`
            element.append(p)
        }

        return element
    },
}

const chatLog = {
    id: 'chat-log',
    type: 'div',
    classList: ['chat-log'],
    create: function () {
        const element = document.createElement(this.type)
        this.classList.forEach((c) => {
            element.classList.add(c)
        })
        element.id = this.id
        return element
    },
    addMessage: function (message) {
        const messageElement = chatMessage.create(
            message.text,
            message.from || message.user_name || message.name
        )
        const element = document.getElementById(this.id)
        element.prepend(messageElement)
    },
}

const chatInput = {
    zIndex: 2000,
    id: 'chat-form',
    inputID: 'message',
    templateID: 'chat-input-template',
    type: 'text-area',
    create: function () {
        const template = document.getElementById(this.templateID)
        const clone = template.content.firstElementChild.cloneNode(true)
        clone.style.zIndex = this.zIndex
        clone.setAttribute('autocomplete', 'off')
        clone.id = this.id
        return clone
    },
    getElement() {
        return document.getElementById(this.id)
    },
    getInputElement() {
        return document.getElementById(this.inputID)
    },
}

const chat = {
    id: 'chat',
    type: 'div',
    classList: ['chat'],
    root: 'root',
    create: function () {
        const element = document.createElement(this.type)
        this.classList.forEach((c) => {
            element.classList.add(c)
        })
        return element
    },
}

const videoAreaOverlay = {
    zIndex: 1000,
    id: 'video-area-overlay',
    type: 'div',
    classList: ['video-area-overlay'],
    create: function () {
        const element = document.createElement(this.type)
        this.classList.forEach((c) => {
            element.classList.add(c)
        })
        element.id = this.id
        element.style.zIndex = this.zIndex
        return element
    },
    addVideo: function (stream) {
        console.log({ stream })
        const videoElement = video.create()
        videoElement.srcObject = stream
        const element = document.getElementById(this.id)
        element.append(videoElement)
        return videoElement
    },
}

const video = {
    type: 'video',
    classList: ['video'],
    attributes: {
        width: '280',
        height: '210',
        autoplay: 'true',
        controls: 'true',
    },
    create: function () {
        const element = document.createElement(this.type)
        this.classList.forEach((c) => {
            element.classList.add(c)
        })
        for (const [key, val] of Object.entries(this.attributes)) {
            element.setAttribute(key, val)
        }
        element.muted = true
        return element
    },
}
const render = () => {
    // Render chat area
    const chatRoot = document.getElementById(chat.root)
    const chatElement = chat.create()
    chatRoot.append(chatElement)

    // render chat log
    const chatLogElement = chatLog.create()
    chatElement.append(chatLogElement)

    // render chat input box
    const chatInputElement = chatInput.create()
    chatElement.append(chatInputElement)

    // add overlay to contain video streams
    const videoArea = videoAreaOverlay.create()
    chatRoot.append(videoArea)

    return {
        chatLog,
        videoAreaOverlay,
        chatInput,
    }
}

export default render
