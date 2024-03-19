const chatMessage = {
    type: 'div',
    classList: ['chat-message'],
    create: function (text, from) {
        const textSpan = document.createElement('span')
        textSpan.innerText = text

        const fromSpan = document.createElement('span')
        fromSpan.classList.add('bold')
        fromSpan.innerText = `${from}: `

        const element = document.createElement(this.type)
        this.classList.forEach((c) => {
            element.classList.add(c)
        })

        element.appendChild(fromSpan)
        element.appendChild(textSpan)

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

const container = {
    id: 'container',
    type: 'div',
    classList: ['container'],
    root: 'root',
    create: function () {
        const element = document.createElement(this.type)
        this.classList.forEach((c) => {
            element.classList.add(c)
        })
        return element
    },
}

const videoArea = {
    zIndex: 1000,
    id: 'video-area',
    type: 'div',
    classList: ['video-area'],
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
        const videoWrapperElement = video.create()

        const videoElement = videoWrapperElement.firstChild
        videoElement.srcObject = stream

        const element = document.getElementById(this.id)
        element.append(videoWrapperElement)

        return videoElement
    },
    removeRemote: function () {
        const children = Array.from(document.getElementById(this.id).children)
        if (children.length > 1) {
            let i = 1
            while (i < children.length) {
                if (children[i]) {
                    children[i].remove()
                    i++
                }
            }
        }
    },
    identifyStream: function (streamID, name) {
        const videoElements = Array.from(
            document.getElementsByClassName('video')
        )
        videoElements.forEach((v) => {
            if (v.srcObject.id === streamID) {
                const textElement = document.createElement('div')
                textElement.classList.add('text-overlay')

                const text = document.createTextNode(name)
                textElement.appendChild(text)

                const wrapper = v.parentElement
                // Remove existing overlay
                Array.from(wrapper.children).forEach((c) => {
                    if (c.classList.contains('text-overlay')) {
                        c.remove()
                    }
                })

                wrapper.appendChild(textElement)
            }
        })
    },
}

const video = {
    type: 'video',
    classList: ['video'],
    attributes: {
        autoplay: 'true',
        controls: 'true',
    },
    create: function () {
        const wrapperElement = document.createElement('div')
        wrapperElement.classList.add('video-wrapper')

        const element = document.createElement(this.type)
        this.classList.forEach((c) => {
            element.classList.add(c)
        })
        for (const [key, val] of Object.entries(this.attributes)) {
            element.setAttribute(key, val)
        }
        element.muted = true
        wrapperElement.appendChild(element)
        return wrapperElement
    },
}
const render = () => {
    const root = document.getElementById(container.root)

    // Render chat area
    const containerElement = container.create()
    root.append(containerElement)

    // add area to contain video streams
    const videoAreaElement = videoArea.create()
    containerElement.append(videoAreaElement)

    // render chat log
    const chatLogElement = chatLog.create()
    containerElement.append(chatLogElement)

    // render chat input box
    const chatInputElement = chatInput.create()
    containerElement.append(chatInputElement)

    return {
        chatLog,
        videoArea,
        chatInput,
    }
}

export default render
