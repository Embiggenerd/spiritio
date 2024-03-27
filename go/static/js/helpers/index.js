/**
 * @argument {import("../../types.js").CommandConfigs} cmdCfg
 * */
export const getElemsFromCmdCfg = (cmdCfg) => {
    const elems = []

    const workOrderKeys = Object.keys(cmdCfg)
    let i = 0
    while (i < workOrderKeys.length) {
        /** @type {import("../../types.js").Elem} */
        const elem = {}
        elem.type = 'div'
        elem.attributes = []
        elem.children = [workOrderKeys[i]]
        elems.push(elem)
        i++
    }
    return elems
}

/** @type {import("../../types.js").getAliasesFromCmdCfg} */
export const getAliasesFromCmdCfg = (cmdCfg) => {
    const aliases = []
    const commandKeys = Object.keys(cmdCfg)
    let i = 0

    while (i < commandKeys.length) {
        const cmd = cmdCfg[commandKeys[i]]
        if (cmd.aliases && cmd.aliases.length > 0) {
            let j = 0
            while (j < cmd.aliases.length) {
                const alias = cmd.aliases[j]
                aliases.push(alias)
                j++
            }
        }
        i++
    }
    return aliases
}
