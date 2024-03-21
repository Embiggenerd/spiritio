/**
 * @type {import("../../types").ChatMessageHelper}
 */
const chatMessage = {
    id: 'chatMessage',
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

/**
 * @type {import("../../types").ChatLogHelper}
 */
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
        if (element) element.prepend(messageElement)
    },
}

/**
 * @type {import("../../types").ChatInputHelper}
 */
const chatInput = {
    classList: [],
    zIndex: '2000',
    id: 'chat-form',
    inputID: 'message',
    templateID: 'chat-input-template',
    type: 'text-area',
    create: function () {
        let template = document.getElementById(this.templateID)
        let assertTemplate = /** @type {HTMLTemplateElement | null} */ (
            template
        )
        let clone
        if (template && assertTemplate) {
            if (assertTemplate.content.firstElementChild)
                clone = assertTemplate.content.firstElementChild.cloneNode(true)
        }

        let assertCloneElement = /** @type {HTMLTemplateElement | null} */ (
            clone
        )

        if (assertCloneElement) {
            assertCloneElement.style.zIndex = this.zIndex
            assertCloneElement.setAttribute('autocomplete', 'off')
            assertCloneElement.id = this.id
        }

        return assertCloneElement
    },
    getElement() {
        return document.getElementById(this.id)
    },
    getInputElement() {
        return document.getElementById(this.inputID)
    },
}

/**
 * @type {import("../../types").ContainerHelper}
 */
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

/**
 * @type {import("../../types").VideoAreaHelper}
 */
const videoArea = {
    // zIndex: 1000,
    id: 'video-area',
    type: 'div',
    classList: ['video-area'],
    create: function () {
        const element = document.createElement(this.type)
        this.classList.forEach((c) => {
            element.classList.add(c)
        })
        element.id = this.id
        // element.style.zIndex = this.zIndex
        return element
    },
    addVideo: function (stream) {
        const videoWrapperElement = video.create()

        const videoElem = videoWrapperElement.firstChild
        const assertVideoMediaElem = /** @type {HTMLMediaElement | null} */ (
            videoElem
        )
        if (assertVideoMediaElem) assertVideoMediaElem.srcObject = stream

        const element = document.getElementById(this.id)
        if (element) element.append(videoWrapperElement)

        return assertVideoMediaElem
    },
    removeRemote: function () {
        const element = document.getElementById(this.id)
        let nodeChildren
        if (element) {
            nodeChildren = element.children
        }
        const children = Array.from(nodeChildren || [])
        const localVideoIndex = 0
        let i = 0
        while (i < children.length) {
            if (i !== localVideoIndex) {
                children[i].remove()
                i++
            }
        }
    },
    identifyStream: function (streamID, name) {
        const videoElements = Array.from(
            document.getElementsByClassName('video')
        )
        const assertVideoMediaElems =
            /** @type {HTMLMediaElement[] | null[]} */ (videoElements)
        assertVideoMediaElems.forEach((v) => {
            if (v && v.srcObject) {
                const assertSrcObjectMediaStream = /** @type {MediaStream} */ (
                    v.srcObject
                )
                if (assertSrcObjectMediaStream.id === streamID) {
                    const textElement = document.createElement('div')
                    textElement.classList.add('text-overlay')

                    const text = document.createTextNode(name)
                    textElement.appendChild(text)

                    const wrapper = v.parentElement
                    let nodeChildren
                    if (wrapper) {
                        nodeChildren = wrapper.children
                    }
                    const children = Array.from(nodeChildren || [])
                    let i = 0
                    while (i < children.length) {
                        if (children[i].classList) {
                            children[i].remove()
                        }
                    }
                }
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
        const assertMediaElement = /** @type {HTMLMediaElement} */ (element)
        this.classList.forEach((c) => {
            assertMediaElement.classList.add(c)
        })
        for (const [key, val] of Object.entries(this.attributes)) {
            assertMediaElement.setAttribute(key, val)
        }
        assertMediaElement.muted = true
        wrapperElement.appendChild(assertMediaElement)
        return wrapperElement
    },
}

/**
 * @type {import("../../types").Render}
 */
const render = () => {
    const root = document.getElementById(container.root)
    if (!root) {
        throw new Error('no root to latch onto')
    }
    // Render chat area
    const containerElement = container.create()
    if (!containerElement) {
        throw new Error('failed to create container element')
    }
    root.append(containerElement)

    // add area to contain video streams
    const videoAreaElement = videoArea.create()
    if (videoAreaElement) containerElement.append(videoAreaElement)

    // render chat log
    const chatLogElement = chatLog.create()
    if (chatLogElement) containerElement.append(chatLogElement)

    // render chat input box
    const chatInputElement = chatInput.create()
    if (chatInputElement) containerElement.append(chatInputElement)

    return {
        chatLog,
        videoArea,
        chatInput,
    }
}

export default render
