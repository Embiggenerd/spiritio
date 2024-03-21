export type Message = {
    type: string
    data: Event | Question
}

export type Event = {
    event: string
    data: any
}

export type Question = {
    ask: string
}

type MessageService = {
    init: () => MessageService
    path: string
    scheme: string
    conn: any
    connect: () => any
    sendMessage: (message: any) => void
    assignCallbacks: (
        handleOpen: any,
        handleError: any,
        handleMessage: any,
        handleClose: any
    ) => void
}

type MediaService = {
    init: () => Promise<MediaService>
    permissionsGranted: boolean
    peerConnection: RTCPeerConnection | null
    constraints: {
        video: boolean
        audio: boolean
    }
    stream: MediaStream | null
    closePeerConnection: () => void
    createAnswer: () => Promise<RTCSessionDescriptionInit> | undefined
    addCandidate: (candidate: RTCIceCandidateInit) => void
    setRemoteDescription: (offer: RTCSessionDescriptionInit) => void
    addTrack: () => void
    assignCallbacks: (trackHandler: any, iceCandidateHandler: any) => void
    setLocalDescription: (
        answer: RTCLocalSessionDescriptionInit | undefined
    ) => void
}

type Render = () => Renderer

type Renderer = {
    chatLog: any
    videoArea: any
    chatInput: any
}

type ChatMessageHelper = ElementHelper & {
    type: string
    classList: string[]
    create: (text: string, form: string) => HTMLElement
}

export type ChatLogHelper = ElementHelper & {
    id: string
    type: string
    classList: string[]
    create: () => HTMLElement
    addMessage: (message: any) => void
}

export type ChatInputHelper = ElementHelper & {
    zIndex: string
    inputID: string
    templateID: string
    getElement: () => HTMLElement | null
    getInputElement: () => HTMLElement | null
}

export type ContainerHelper = ElementHelper & {
    root: string
}

export type VideoAreaHelper = ElementHelper & {
    addVideo: (stream: MediaStream) => HTMLMediaElement | null
    removeRemote: () => void
    identifyStream: (streamID: string, name: string) => void
}

export type VideoHelper = ElementHelper & {
    attributes: any
}

export type Component = {
    init: (
        render: Render,
        messageService: MessageService,
        mediaService: MediaService,
        parse: () => Parser
    ) => void
    renderer: Renderer | null
    mediaService: MediaService | null
    messageService: MessageService | null
    parse: null | (() => Parser)
    assignHandleChatInput: () => void
    handleOpen: () => void
    handleClose: () => void
    handleError: (error: any) => void
    onChatSubmit: (event: any) => void
    setOnChatKeyDown: (chatInputElement: HTMLElement) => void
    getCommandLog: () => string[]
    addToCommandLog: (command: string) => void
    orderWork: (work: WorkOrder) => void
    handleEvent: (event: string, data: any) => void
    handleIceCandidate: (e: any) => void
    handleMessage: (event: any) => void
    handleQuestion: (ask: string) => void
    handleMessageError: (event: any) => void
    orderMedia: () => void
    handleOnTrack: (event: any) => void
}

export type CommandConfig = {
    workOrder: string
    args: commandArgument[]
}

export type WorkOrder = {
    order: string
    details: any
}

export interface CommandConfigs {
    [key: string]: CommandConfig
}

export type Parser = {
    command: string
    i: number
    workOrderKey: string
    commandConfigs: CommandConfigs
    commandChar: string
    allLettersRegex: RegExp
    parseUserCommand: (command: string) => CommandConfig
    parse: () => void
    match: (str: string) => boolean
    eat: (str: string) => void
    readWhileMatching: (regex: RegExp) => string
    skipWhitespace: () => void
    parseCommand: (word: string) => void
    parseArguments: () => void
}

// export type CommandConfigs = Record<Student['id'], Student>

type commandArgument = {
    name: string
    regex: RegExp
    value: string
}

type ElementHelper = {
    id: string
    type: string
    classList: string[]
    create: () => HTMLElement | null
}
