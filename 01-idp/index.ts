import * as pulumi from "@pulumi/pulumi";
import {ArgoCD} from "./argocd";

let initialObjects = new pulumi.asset.FileAsset("./argocd-initial-objects.yaml");

const argocd = new ArgoCD("argocd", {
    initialObjects: initialObjects,
    version: "7.7.0",
})
