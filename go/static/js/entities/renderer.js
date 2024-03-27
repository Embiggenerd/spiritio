// import commandConfigs from '../config/commandConfig.js'
import commandsConfig from '../config/commandConfig.js'
import { getElemsFromCmdCfg } from '../helpers/index.js'

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
 * @type {import("../../types").ChatLog}
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
            message.from_user_name
        )
        const element = document.getElementById(this.id)
        if (element) element.prepend(messageElement)
    },
}

/** @type {import("../../types.js").Tooltip} */
const tooltip = {
    tooltipID: 'tooltip',
    tooltipNameClass: 'tooltip-name',
    templateID: 'tooltip-template',
    element: null,
    chatInputElement: null,
    create: function (chatInputElement) {
        this.chatInputElement = chatInputElement
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

        this.element = assertCloneElement

        if (assertCloneElement) {
            assertCloneElement.setAttribute('autocomplete', 'off')
        }

        return assertCloneElement
    },
    hide() {
        const tooltip = this.getElement()
        if (tooltip) tooltip.classList.add('hidden')
    },
    show() {
        const tooltip = this.getElement()
        if (tooltip) tooltip.classList.remove('hidden')
    },

    setContent(elems) {
        let i = 0
        const tooltip = this.getElement()
        if (tooltip) tooltip.innerHTML = ''

        while (i < elems.length && tooltip) {
            let str = ''
            const element = document.createElement(elems[i].type)

            let j = 0
            while (j < elems[i].children.length) {
                const child = elems[i].children[j]
                if (typeof child === 'string') {
                    str = `${child} `
                    element.innerText = str
                }
                j++
            }

            // Storing id and name to translate name to id in direct message command
            let k = 0
            while (k < elems[i].attributes.length) {
                const attr = elems[i].attributes[k]
                const name = attr.name
                const value = attr.value
                element.setAttribute(name, value)
                k++
            }

            element.addEventListener('click', () => {
                const input = this.getInputElement()
                if (input) {
                    const symbol = input.value[0]
                    input.value = symbol + str
                    input.focus()
                    setTimeout(() => {
                        input.setSelectionRange(
                            input.value.length,
                            input.value.length
                        )
                    }, 0)
                }
            })

            tooltip.appendChild(element)
            i++
        }
    },
    getElemsData() {
        const tt = this.getElement()
        let elements
        if (tt) {
            elements = Array.from(tt.children)
            let i = 0
            const elems = []
            while (i < elements.length) {
                const e = elements[i]
                if (e instanceof HTMLElement) {
                    elems.push({
                        type: e.tagName.toLocaleLowerCase(),
                        children: [e.dataset['name'] || ''],
                        attributes: Array.from(e.attributes).map(
                            ({ name, value }) => ({ name, value })
                        ),
                    })
                }
                i++
            }
            return elems
        }
    },
    getElement() {
        return this.element
    },
    getInputElement() {
        return this.chatInputElement
    },
}

/**
 * @type {import("../../types").ChatForm}
 */
const chatForm = {
    classList: [],
    zIndex: '2000',
    id: 'chat-form',
    inputID: 'message',
    templateID: 'chat-input-template',
    tooltipNames: { ...tooltip },
    tooltipCommands: { ...tooltip },
    tooltipNameClass: 'tooltip-name',
    type: 'form',
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
        const elem = document.getElementById(this.inputID)
        return /** @type {HTMLInputElement} */ (elem)
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
    id: 'video-area',
    type: 'div',
    classList: ['video-area'],
    create: function () {
        const element = document.createElement(this.type)
        this.classList.forEach((c) => {
            element.classList.add(c)
        })
        element.id = this.id
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
        let i = 1
        while (i < children.length) {
            if (children[i]) {
                children[i].remove()
                i++
            }
        }
    },
    identifyStream: function (streamID, name) {
        const videoElements = Array.from(
            document.getElementsByClassName('video')
        )

        if (streamID === '') {
            const textElement = this.createTextOverlay(name)

            const wrapper = videoElements[0].parentElement
            wrapper?.appendChild(textElement)
        }

        const assertVideoMediaElems =
            /** @type {HTMLMediaElement[] | null[]} */ (videoElements)
        assertVideoMediaElems.forEach((v) => {
            if (v && v.srcObject) {
                const assertSrcObjectMediaStream = /** @type {MediaStream} */ (
                    v.srcObject
                )
                if (assertSrcObjectMediaStream.id === streamID) {
                    const textElement = this.createTextOverlay(name)

                    const wrapper = v.parentElement
                    let nodeChildren

                    if (wrapper) {
                        nodeChildren = wrapper.children
                    }

                    const children = Array.from(nodeChildren || [])
                    children.forEach((c) => {
                        if (c.classList.contains('text-overlay')) {
                            c.remove()
                        }
                    })

                    wrapper?.appendChild(textElement)
                }
            }
        })
    },

    createTextOverlay: function (name) {
        const textElement = document.createElement('div')
        textElement.classList.add('text-overlay')

        const textNode = document.createTextNode(name)
        textElement.appendChild(textNode)
        return textElement
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

    // Add area to contain video streams
    const videoAreaElement = videoArea.create()
    if (videoAreaElement) containerElement.append(videoAreaElement)

    // Render chat log
    const chatLogElement = chatLog.create()
    if (chatLogElement) containerElement.append(chatLogElement)

    // Render chat input box
    const chatInputElement = chatForm.create()
    if (chatInputElement) {
        containerElement.append(chatInputElement)

        const tooltipCommands = chatForm.tooltipCommands.create(
            chatForm.getInputElement()
        )
        const assertTtCmd = /** @type {Node} */ (tooltipCommands)
        chatInputElement.appendChild(assertTtCmd)

        chatForm.tooltipCommands.setContent(getElemsFromCmdCfg(commandsConfig))

        const tooltipNames = chatForm.tooltipNames.create(
            chatForm.getInputElement()
        )

        const assertTtNamesNode = /** @type {Node} */ (tooltipNames)

        chatInputElement.appendChild(assertTtNamesNode)
    }

    return {
        chatLog,
        videoArea,
        chatForm,
    }
}

export default render
