/**
 * Book Library
 * This is a sample API that describes the structure of our Book-Library-Server 
 *
 * The version of the OpenAPI document: 1.0.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */
import { SessionState } from './sessionState';
import { UserMessage } from './userMessage';


export interface ErrorResponse { 
    sessionState?: SessionState;
    psuMessages?: Array<UserMessage>;
}

