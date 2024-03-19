const media = {
    peerConnection: null,
    constraints: {
        video: true,
        audio: true,
    },
    stream: null,
    async init() {
        try {
            // Ask for mic and cam access, set media stream
            this.stream = await navigator.mediaDevices.getUserMedia(
                this.constraints
            )
            this.peerConnection = new RTCPeerConnection()
            return this
        } catch (e) {
            // If rejected, return null
            return null
        } 
    },
    closePeerConnection: function () {
        this.stream.getTracks().forEach((s) => {
            s.stop()
        })
    },
    createAnswer: function () {
        return this.peerConnection.createAnswer()
    },
    addCandidate: function (candidate) {
        this.peerConnection.addIceCandidate(candidate)
    },
    setRemoteDescription: function (offer) {
        this.peerConnection.setRemoteDescription(offer)
    },
    addTrack: function () {
        this.stream.getTracks().forEach((track) => {
            this.peerConnection.addTrack(track, this.stream)
        })
    },
    assignCallbacks: function (trackHandler, iceCandidateHandler) {
        this.peerConnection.ontrack = trackHandler
        this.peerConnection.onicecandidate = iceCandidateHandler
    },
    setLocalDescription: function (answer) {
        this.peerConnection.setLocalDescription(answer)
    },
}

export default media
