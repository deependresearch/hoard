# HOARD
Historical Observations of Actionable Reputation Data

Components:
```
> Golang
> Hi Speed RegEx Matching
> Cuckoo Filters
> Graylog/Elastic Search/Splunk?
> Threat Exchanges (Alienvault/Anomali/ThreatConnect/XForce)
```

The RegEx's can consume IPv4/IPv6 internal IPs and Domains, the observable wont take much per filter.  5GB Flat text drops to <200M of Index'd data. That turns 20+GB/Day into <1GB of index data in an observable rich environment.

## This project depends upon:

Go Package Dependencies:
```
    > "github.com/seiflotfy/cuckoofilter" // MIT License
    > "github.com/go-redis/redis" // BSD 2 License
```

Applications (not build within the program itself):
```
    > Redis
    > Graylog, ELK or Splunk  # Not yet implemented.
    > RSyslog omhiredis  # Functional Support in POC
    > Suricata  # Functional Support in POC
```


#  External Contributions (WELCOME!)

This was my first project with multiple moving parts in GoLang. I'm sure to have bugs or areas that need guidance.
If you have an idea, found a bug or want to help us be more effecient with our code, please submit a pull request!


#  Background

>> This product is currently a *Proof Of Concept*; I'm sure we aren't the first to consider it and its probably not the best solution. I heard about a reputable vendor doing something similar using in memory high effeciency databases and sha hashing. That sounded really cool too!
>> Community participation is encouraged!

Data reputation systems classify observations made and recognized by machines such as Internet Protocol (IP) addresses, Domain names, email addresses, and computed hashes either fuzzy matching (those known as SSDEEP) or fixed hashes computed with algorithms known as MD5 or SHA digesting.

Observable data is frequently identified by computer security devices, intrusion detection systems, and forensic investigators following an intrusion or other malicious event. When observable data is paired with contextual information it becomes known as an indicator. Indicators are usually given a reputation or risk score.

These Indicators are frequently classified with the industry term "threat Intelligence" and disseminated by both computers and machines to alert computer security teams about threats they may have been previously unaware of. STIX is the standard format by which these observables are shared.

Many computer security technologies will import this threat intelligence data and match it with same type observables. This has been done through Security Incident and Event Monitoring (SIEM) solutions, antivirus and network or system level intrusion detection systems.

The observations may be shared manually or programmatically; however, they suffer from ephemeral challenges. Once identified, an adversary may change their attack profile and in doing so they change the identified observables. This asserts that even the fastest sharing platforms are likely to become less effective in the hours to days following the initial discovery of a given observable.

For organizations which capture log traffic, searching in a reverse chronological order may provide additional value, allowing an organization to determine if they have been affected by a situation.

These searches are time consuming and result in matching known malicious observables with hundreds of gigabytes or terabytes of data.

# Abstract
This application aims to reduce the speed and storage limitations needed for quickly matching observable data with cataloged threats.

Using a system of pattern matching, known as Regular Expression searches, an analyst can quickly update observables they wish to collect from their security systems. Regular Expressions are regularly used in existing security devices which makes the system ideal for this solution; but JSON formatted output from devices like Suricata are easily parsed.

Once identified, an application will monitor log events in real time, passing the log event to a queuing system which will store the data temporarily while another application extracts and stores only the observable data identified by the analyst as being relevant. This immediately reduces the data stored to a manageable size and provides raw data that can be indexed in a probabilistic data structure known as sketches or Cuckoo Filters.

This data structure allows for the storage of large volumes of data in such a way that it can be quickly searched. At predetermined time intervals, the sketches (Cuckoo Filters) will be written to disk using a naming convention that identifies its date and time stamps. These sketches can then be stored indefinitely for later searches, either in databases or directly on disk alongside the original logs.

Once observable data has been added to a threat intelligence exchange platform or a security team has been alerted to an issue by law enforcement, security researchers or media releases, a second application can be utilized to rapidly search back in time by rapidly querying the sketches to determine a probability rating that the observable had been cataloged by the security device.

When a probabilistic match has been identified, the organizations Security Incident and Event Monitor (SIEM) is queried using the file date and timestamp information. Because the searches are scoped with information including the date, time and observable, the search will run much faster. The results from the SIEM searches will give security analysts additional information about the threat and give the analyst the capability to match the threat with the host that generated the activity in question.

Note: Since the observations are "probablistic" in nature, a second search within the raw data store (SIEM) will be required to validate the observation.

# Spec
RSYSLOG Is an application developed under the GNU license structure and is a common technology used to capture event logs from security devices, system state monitors and other computer components and services.

RSYSLOG has a component known as "omhiredis" which outputs log events to a Redis Queuing system using LPUSH. Our Application, known as HOARD Server will then pull data off the Redis queue using a technique known as RPOP.

Suricata has similar built in functionality for EVE Format. Since this format is single line JSON, we can rapidly parse for interesting artifacts without requiring expensive regex parsing.

HOARD will search each log event line for regular expressions that match analyst-controlled values for IP address, Domain Names, Email Addresses and other observables. This observable data will be written to back to Redis queues in anticipation of building a Cuckoo Filter/Sketch. At configurable time intervals (2 hours by default) these sketches will be written to disk for later retrieval and matching by a second HOARD application (HOARD Client). Events or areas of the log event that do not match the regular expression type provided will be dropped or ignored.

This matching methodology will allow for expansion of this technology into other areas where rapid searches are required, such as fraud identification and data leak prevention. With example log events at 5GB in size, extracted cuckoo filter sketches are on average 25+ times smaller, making them easy to transport and perform further offline analysis, including independent third parties (forensic teams) who have access to broader data sets.

The Cuckoo library created by Seif Lotfy (seiflotfy) is extremely fast, searching over a weeks worth of data in mere seconds. By Pre-Searching for probable timeframes we reduce load on the SIEM.

Events shared through threat intelligence exchanges, distributed by electronic mail or delivered by another media may be rapidly searched using the HOARD Client Search application which will quickly query all sketches and eventually reach out and query third party SIEM solutions to validate the match.

By using cuckoo filter sketches for searching, the application is able to rapidly search through high volumes of data and provide a probabilistic match, which gives a logic map of the observable combined with the date and time of the event. This data can be programmatically or manually searched in the organizations log collection platform to provide additional context such as source or originating device and other associated behaviors that may or may not be known.

Following the enrichment, details may be sent to Security or Technical operations teams for triage and additional analysis if warranted.

Hoard writes the output files with fhe following format, which makes it easy to narrow scope or share with third party analysts, while not providing all private company data to the third party organization:

HOARD_index_YYYYMMDD_HHMM-YYYYMMDD_HHMM.cf -- This represents a FROM and TO datestamp that can be used to help narrow scope in the SIEM or scan for an event duing a pre-defined window such as a pentest or suspected incident.

For environments with less activity, we write a maximum of every two hours. In high-volume enviorments we will write more frequently and SIEM searches will be better scoped. This allows us to grow with organizational need.

We should make sure that we have a minimum of four processes pulling data off the queue. Its likely that more will be required in bigger environments.

Whenever possible, we'd like to run a unique instance of HOARD Storage for each SIEM logger (Graylog/Splunk/etc) rather than try to be a consumer for everything.

Process breakdown (Daemon):
    1 process to handle general setup/application Tasks
    4-8 processes to handle pulling data off Redis Queues and performing REGEX /Cuckoo searches.
    1 process to count the filter regularly and write observables

One process as a gate to make sure we're pulling in data from Redis. If the counts drop below XX EPS in Redis we may wish to take action.

Notes:  Our code is not crash-safe, there is a potential we could lose the data we've already accumulated while parsing, this could be up to 10k events. For this reason, we will use Redis again to write our filters into a temporary queue; which can be configured to write those queues to disk periodically, thus forcing our application to do less heavy lifting.


# Suricata Config:

```
  - eve-log:
      enabled: yes
      type: redis #file|syslog|unix_dgram|unix_stream
      redis:
        enabled: yes
        server: 127.0.0.1
        port: 6379
        async: true
        mode: list
        key: suricata
      types:
        - dns
        - http
```

# Product Phases

## PHASE 1:
    Propose Project, document idea and workflow
    Build working POC: Suricata Log -> Send to Redis (PUSH/POP QUEUE), Parse by HOARD.
    License and release publicly (est: July 2018)
    * COMPLETE *

## PHASE 2:
    Add additional log types (non-JSON Structured) for RSYSLOG/etc.  * COMPLETE *
    Add file-based tailer for environments where log files are not written to SIEM (Suricata Example Done)
        This way raw logs can be stored "off machine" where disk space isn't a premium but the filters can still be used to search.
    Consider using a publish/subscribe queue so that we can integrate with other solutions at larger instituions.

## PHASE 3:
    Implement Searching in Graylog/ELK/Splunk and Slack or Email based alerting.
        May consider writing a local file to the system and indexing with SIEM, which could do subsearching and alerting?
    Implement STIX Parser for consuming data in industry standard format.
    Develop installer and daemon mode/system init scripts.
    Build Docker container for easier deployment to some environments

