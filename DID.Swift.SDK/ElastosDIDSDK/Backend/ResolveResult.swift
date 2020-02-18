import Foundation

class ResolveResult {
    private var _did: DID
    private var _status: ResolveResultStatus
    private var _idtransactionInfos: Array<IDTransactionInfo>?
    
    init(_ did: DID, _ status: Int) {
        self._did = did
        self._status = ResolveResultStatus(rawValue: status)!
    }

    var did: DID {
        return self._did
    }

    var status: ResolveResultStatus {
        return self._status
    }

    var transactionCount: Int {
        return self._idtransactionInfos?.count ?? 0
    }

    func transactionInfo(_ index: Int) -> IDTransactionInfo? {
        return self._idtransactionInfos?[index]
    }

    func appendTransactionInfo( _ info: IDTransactionInfo) { // TODO: should be synchronized ?
        if  self._idtransactionInfos == nil {
            self._idtransactionInfos = Array<IDTransactionInfo>()
        }
        self._idtransactionInfos!.append(info)
    }

    class func fromJson(_ node: JsonNode) throws -> ResolveResult {
        guard !node.isEmpty else {
            throw DIDError.illegalArgument()
        }

        let error = { (des: String) -> DIDError in
            return DIDError.didResolveError(des)
        }
        let serializer = JsonSerializer(node)
        var options: JsonSerializer.Options

        options = JsonSerializer.Options()
                                .withHint("resolved result did")
                                .withError(error)
        let did = try serializer.getDID(Constants.DID, options)

        options = JsonSerializer.Options()
                                .withRef(-1)
                                .withHint("resolved status")
                                .withError(error)
        let status = try serializer.getInteger(Constants.STATUS, options)

        let result = ResolveResult(did, status)
        if status != ResolveResultStatus.STATUS_NOT_FOUND.rawValue {
            let transactions = node.getNodeArray(Constants.TRANSACTION)
            guard transactions?.count ?? 0 > 0 else {
                throw DIDError.didResolveError("invalid resolve result.")
            }
            for transaction in transactions! {
                result.appendTransactionInfo(try IDTransactionInfo.fromJson(transaction))
            }
        }
        return result
    }

    class func fromJson(_ json: Data) throws -> ResolveResult {
        guard !json.isEmpty else {
            throw DIDError.illegalArgument()
        }

        let node: Dictionary<String, Any>?
        do {
            node = try JSONSerialization.jsonObject(with: json, options: []) as? Dictionary<String, Any>
        } catch {
            throw DIDError.didResolveError("Parse resolve result error")
        }
        return try fromJson(JsonNode(node!))
    }

    class func fromJson(_ json: String) throws -> ResolveResult {
        return try fromJson(json.data(using: .utf8)!)
    }

    private func toJson(_ generator: JsonGenerator) {
        generator.writeStartObject()

        generator.writeStringField(Constants.DID, self.did.toString())
        generator.writeNumberField(Constants.STATUS, self.status.rawValue)

        if (self._status != .STATUS_NOT_FOUND) {
            generator.writeFieldName(Constants.TRANSACTION)
            generator.writeStartArray()

            for txInfo in self._idtransactionInfos! {
                txInfo.toJson(generator)
            }
            generator.writeEndArray()
        }

        generator.writeEndObject()
    }

    func toJson() throws -> String {
        // TODO
        return "TODO"
    }
}

extension ResolveResult: CustomStringConvertible {
    @objc public var description: String {
        return (try? toJson()) ?? ""
    }
}
