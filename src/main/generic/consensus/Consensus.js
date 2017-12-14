class Consensus {
    /**
     * @param {NetworkConfig} [netconfig]
     * @returns {Promise.<FullConsensus>}
     */
    static async full(netconfig) {
        await Crypto.prepareSyncCryptoWorker();

        /** @type {ConsensusDB} */
        const db = await ConsensusDB.getFull();
        /** @type {Accounts} */
        const accounts = await Accounts.getPersistent(db);
        /** @type {FullChain} */
        const blockchain = await FullChain.getPersistent(db, accounts);
        /** @type {Mempool} */
        const mempool = new Mempool(blockchain, accounts);
        /** @type {NetworkConfig} */
        const networkConfig = netconfig || new NetworkConfig();
        /** @type {Network} */
        const network = await new Network(blockchain, networkConfig);

        return new FullConsensus(blockchain, mempool, network);
    }

    /**
     * @param {NetworkConfig} [netconfig]
     * @returns {Promise.<LightConsensus>}
     */
    static async light(netconfig) {
        await Crypto.prepareSyncCryptoWorker();

        /** @type {ConsensusDB} */
        const db = await ConsensusDB.getLight();
        /** @type {Accounts} */
        const accounts = await Accounts.getPersistent(db);
        /** @type {LightChain} */
        const blockchain = await LightChain.getPersistent(db, accounts);
        /** @type {Mempool} */
        const mempool = new Mempool(blockchain, accounts);
        /** @type {NetworkConfig} */
        const networkConfig = netconfig || new NetworkConfig(new Services(Services.LIGHT, Services.LIGHT | Services.FULL));
        /** @type {Network} */
        const network = await new Network(blockchain, networkConfig);

        return new LightConsensus(blockchain, mempool, network);
    }

    /**
     * @param {NetworkConfig} [netconfig]
     * @returns {Promise.<NanoConsensus>}
     */
    static async nano(netconfig) {
        await Crypto.prepareSyncCryptoWorker();

        /** @type {NanoChain} */
        const blockchain = await new NanoChain();
        /** @type {NanoMempool} */
        const mempool = new NanoMempool();
        /** @type {NetworkConfig} */
        const networkConfig = netconfig || new NetworkConfig(new Services(Services.NANO, Services.NANO | Services.LIGHT | Services.FULL));
        /** @type {Network} */
        const network = await new Network(blockchain, networkConfig);

        return new NanoConsensus(blockchain, mempool, network);
    }
}
Class.register(Consensus);
