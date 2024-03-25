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
            message.from_user_name
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
    tooltipID: 'tooltip',
    tooltipNameClass: 'tooltip-name',
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
        const elem = document.getElementById(this.inputID)
        return /** @type {HTMLInputElement} */ (elem)
    },
    getTooltip() {
        return document.getElementById(this.tooltipID)
    },
    hideTooltip() {
        const tooltip = this.getTooltip()
        if (tooltip) tooltip.classList.add('hidden')
    },
    showTooltip() {
        const tooltip = this.getTooltip()
        if (tooltip) tooltip.classList.remove('hidden')
    },
    setTooltipContent(userArray) {
        let i = 0
        const tooltip = this.getTooltip()
        if (tooltip) tooltip.innerHTML = ''

        while (i < userArray.length && tooltip) {
            const str = `@${userArray[i].name}`

            const text = document.createTextNode(str)
            const elem = document.createElement('div')
            elem.classList.add(this.tooltipNameClass)

            // storing for later use
            elem.dataset.name = userArray[i].name
            elem.dataset.id = userArray[i].id.toString()

            elem.appendChild(text)

            elem.addEventListener('click', () => {
                const input = this.getInputElement()
                input.value = str
                input.focus()
                setTimeout(() => {
                    input.setSelectionRange(
                        input.value.length,
                        input.value.length
                    )
                }, 0)
            })

            tooltip.appendChild(elem)
            i++
        }
    },
    getNamesToIDs() {
        const elems = document.getElementsByClassName(this.tooltipNameClass)
        let i = 0
        const namesToIDs = []
        while (i < elems.length) {
            const e = elems[i]
            if (e instanceof HTMLElement) {
                namesToIDs.push({
                    name: e.dataset.name || '',
                    id: e.dataset.id || '',
                })
            }
            i++
        }
        return namesToIDs
    },
    appendTooltipContent({ name, id }) {
        const currentNames = document.getElementsByClassName(
            this.tooltipNameClass
        )
        let i = 0
        // If name already exists im tooltip, do not append
        while (i < currentNames.length) {
            const assertHTMLElem = /** @type {HTMLElement} */ (currentNames[i])
            if (assertHTMLElem.dataset.id === id) return
            i++
        }

        const tooltip = this.getTooltip()

        const str = `@${name}`

        const text = document.createTextNode(str)
        const elem = document.createElement('div')
        elem.classList.add(this.tooltipNameClass)

        // storing for later use
        elem.dataset.name = name
        elem.dataset.id = id.toString()

        elem.appendChild(text)

        elem.addEventListener('click', () => {
            const input = this.getInputElement()
            input.value = str
            input.focus()
            setTimeout(() => {
                input.setSelectionRange(input.value.length, input.value.length)
            }, 0)
        })

        tooltip?.appendChild(elem)
    },

    removeTooltipContent(id) {
        const currentNames = document.getElementsByClassName(
            this.tooltipNameClass
        )

        let i = 0
        while (i < currentNames.length) {
            const assertHTMLElem = /** @type {HTMLElement} */ (currentNames[i])
            if (assertHTMLElem.dataset.id === id) {
                assertHTMLElem.remove()
            }
            i++
        }
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
