const media = {
    peerConnection: null,
    constraints: {
        video: true,
        audio: true,
    },
    stream: null,
    init: async function () {
        // Ask for mic and cam access, set media stream
        this.stream = await navigator.mediaDevices.getUserMedia(
            this.constraints
        )
        // Create peer connection
        this.peerConnection = new RTCPeerConnection()
        // peerConnection.close(() => {
        //     peerConnection.getReceivers().forEach((track) => {
        //         peerConnection.removeTrack(track)
        //     })
        // })

        return this
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
