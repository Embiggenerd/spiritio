import { getAliasesFromCmdCfg } from './index.js'

/**
 * Parses a user message starting with '/' into a work order
 * @type {import("../../types.js").Parser}
 */
const parser = {
    command: '',
    i: 0,
    workOrderKey: '',
    commandConfigs: {},
    commandChar: '/',
    directMessageChar: '@',
    allLettersRegex: /[a-zA-Z]/,
    alphaNumericSpecialRegex: /[A-Za-z0-9_@./#&+!$_*+-]/,
    aliases: [],

    parseUserCommand: function (command, commandConfigs) {
        this.command = command
        this.commandConfigs = commandConfigs

        this.aliases = getAliasesFromCmdCfg(commandConfigs)

        this.commandConfigs
        this.parse()

        return this.commandConfigs[this.workOrderKey]
    },

    parse: function () {
        if (this.match(this.commandChar)) {
            this.eat(this.commandChar)
        }

        this.parseCommand(this.readWhileMatching(this.allLettersRegex))
        if (!this.commandConfigs.hasOwnProperty(this.workOrderKey)) {
            throw new Error('no such command')
        }

        this.skipWhitespace()
        const argsCount = this.parseArguments(
            this.readWhileMatching(
                this.commandConfigs[this.workOrderKey].args[0].regex
            ),
            0
        )
        if (argsCount !== this.commandConfigs[this.workOrderKey].args.length) {
            throw new Error(
                `wrong argument number: have ${argsCount}, want ${
                    this.commandConfigs[this.workOrderKey].args.length
                }`
            )
        }
    },

    // Find any words that are in the list of possible commands
    parseCommand: function (word) {
        this.aliases.forEach((a) => {
            if (this.match(a)) {
                // If we match to any aliases,
                let workOrderKey = ''
                const keys = Object.keys(this.commandConfigs)
                let i = 0
                while (i < keys.length) {
                    const key = keys[i] // Find which key the alias belongs to
                    if (this.commandConfigs[key].aliases?.includes(a)) {
                        workOrderKey = key
                        break
                    }

                    i++
                }
                this.workOrderKey = workOrderKey // Set our work order key to alias' daddy
                this.eat(a) // Skip past the alias to the arguments and return
                return
            }
        })
        this.workOrderKey = this.workOrderKey + ' ' + word
        this.workOrderKey = this.workOrderKey.trim()

        if (
            Object.keys(this.commandConfigs).includes(
                this.workOrderKey.toLowerCase()
            ) ||
            this.i === this.command.length
        ) {
            return
        }

        this.skipWhitespace()
        this.parseCommand(this.readWhileMatching(this.allLettersRegex))
    },

    parseArguments: function (arg, argIndex) {
        this.skipWhitespace()
        if (!arg) {
            return argIndex
        }

        const argCfg = this.commandConfigs[this.workOrderKey].args[argIndex]
        if (argIndex < this.commandConfigs[this.workOrderKey].args.length) {
            argCfg.value = arg
        }

        const nextArgIndex = argIndex + 1

        let nextRegex
        // Use last regex in case there are extra arguments provided by user
        if (!this.commandConfigs[this.workOrderKey].args[nextArgIndex]) {
            nextRegex =
                this.commandConfigs[this.workOrderKey].args[
                    this.commandConfigs[this.workOrderKey].args.length - 1
                ].regex
        } else {
            nextRegex =
                this.commandConfigs[this.workOrderKey].args[nextArgIndex].regex
        }

        return this.parseArguments(
            this.readWhileMatching(nextRegex),
            nextArgIndex
        )
    },

    parseDirectMessage() {
        this.workOrderKey = 'direct message'
        const name = this.readWhileMatching(this.alphaNumericSpecialRegex)

        this.commandConfigs[this.workOrderKey].args[0].value = name
        this.skipWhitespace()
        const text = this.readWhileMatching(/./)
        this.commandConfigs[this.workOrderKey].args[1].value = text.trimEnd()
    },

    readWhileMatching: function (regex) {
        let startIndex = this.i
        while (
            this.i < this.command.length &&
            regex.test(this.command[this.i])
        ) {
            this.i++
            if (this.i - startIndex > 255) break
        }
        return this.command.slice(startIndex, this.i)
    },

    eat: function (str) {
        if (this.match(str)) {
            this.i += str.length
        } else {
            throw new Error(`Parse error: expecting "${str}"`)
        }
    },

    match: function (str) {
        return this.command.slice(this.i, this.i + str.length) === str
    },

    skipWhitespace: function () {
        this.readWhileMatching(/[\s\n]/)
    },
}

export default () => ({ ...parser })
