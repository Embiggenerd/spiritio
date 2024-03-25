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
    init: (wsClass?: any) => MessageService
    webSocket: any
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
    chatLog: ChatLogHelper
    videoArea: VideoAreaHelper
    chatInput: ChatInputHelper
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
    addMessage: (message: { text: string; from_user_name: string }) => void
}

export type ChatInputHelper = ElementHelper & {
    zIndex: string
    inputID: string
    templateID: string
    tooltipID: string
    tooltipNameClass: string
    getElement: () => HTMLElement | null
    getInputElement: () => HTMLInputElement
    setTooltipContent: (namesToIDs: namesToID[]) => void
    appendTooltipContent: (nameToId: namesToID) => void
    removeTooltipContent: (id: string) => void
    getTooltip: () => HTMLElement | null
    hideTooltip: () => void
    showTooltip: () => void
    getNamesToIDs: () => namesToID[]
}
type namesToID = { name: string; id: string }

export type ContainerHelper = ElementHelper & {
    root: string
}

export type VideoAreaHelper = ElementHelper & {
    addVideo: (stream: MediaStream) => HTMLMediaElement | null
    removeRemote: () => void
    identifyStream: (streamID: string, name: string) => void
    createTextOverlay: (text: string) => HTMLElement
}

export type VideoHelper = ElementHelper & {
    attributes: any
}

export type Component = {
    init: (
        render: Render,
        messageService: MessageService,
        mediaService: MediaService
    ) => void
    renderer: Renderer | null
    mediaService: MediaService | null
    messageService: MessageService | null
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
    handleInput: (event: any) => void
}

export type CommandConfig = {
    workOrder: string
    args: commandArgument[]
}

export type WorkOrder = {
    order: string
    details?: any
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
    directMessageChar: string
    allLettersRegex: RegExp
    alphaNumericSpecialRegex: RegExp
    namesToIDs: namesToID[]
    parseUserCommand: (
        command: string,
        namesToIDs: namesToID[],
        commandConfigs: CommandConfigs
    ) => CommandConfig
    parse: () => void
    match: (str: string) => boolean
    eat: (str: string) => void
    readWhileMatching: (regex: RegExp) => string
    skipWhitespace: () => void
    parseCommand: (word: string) => void
    parseArguments: (arg: string, count: number) => number
    parseDirectMessage: () => void
}

type commandArgument = {
    name: string
    regex: RegExp
    value: string | number
}

type ElementHelper = {
    id: string
    type: string
    classList: string[]
    create: () => HTMLElement | null
}

type UserMessageData = {
    text: string
    from_user_name: string
    from_user_id: number
    user_verified: boolean
    to_user_id: number
}

type UserMessageWorkDetails = {
    text: string
    to_user_id: number | null
}
