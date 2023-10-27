import sys
import operator
from common import common_type as ct

sys.path.append("..")
from utils.util_log import test_log as log


def ip_check(ip):
    if ip == "localhost":
        return True

    if not isinstance(ip, str):
        log.error(f"[IP_CHECK] IP({ip}) is not a string.")
        return False

    return True


def number_check(num):
    if str(num).isdigit():
        return True

    log.error(f"[NUMBER_CHECK] Number({num}) is not a numbers.")
    return False


def exist_check(param, _list):
    if param in _list:
        return True

    log.error(f"[EXIST_CHECK] Param({param}) is not in ({_list}).")
    return False


def dict_equal_check(dict1, dict2):
    if not isinstance(dict1, dict) or not isinstance(dict2, dict):
        log.error(
            f"[DICT_EQUAL_CHECK] Type of dict({str(dict1)}) or dict({str(dict2)}) is not a dict."
        )
        return False
    return operator.eq(dict1, dict2)


def list_de_duplication(_list):
    if not isinstance(_list, list):
        log.error(f"[LIST_DE_DUPLICATION] Type of list({str(_list)}) is not a list.")
        return _list

    # de-duplication of _list
    result = list(set(_list))

    # Keep the order of the elements unchanged
    result.sort(key=_list.index)

    log.debug(
        f"[LIST_DE_DUPLICATION] {str(_list)} after removing the duplicate elements, the list becomes {result}"
    )
    return result


def list_equal_check(param1, param2):
    check_result = True

    _list1 = list_de_duplication(param1)
    _list2 = list_de_duplication(param2)

    if len(_list1) == len(_list2):
        for i in _list1:
            if i not in _list2:
                check_result = False
                break
    else:
        check_result = False
    if check_result is False:
        log.error(
            f"[LIST_EQUAL_CHECK] List({str(param1)}) and list({str(param2)}) are not equal."
        )

    return check_result


def list_contain_check(sublist, superlist):
    if not isinstance(sublist, list):
        raise Exception(f"{sublist} isn't list type")
    if not isinstance(superlist, list):
        raise Exception(f"{superlist} isn't list type")

    check_result = True
    for i in sublist:
        if i not in superlist:
            check_result = False
            break
        else:
            superlist.remove(i)
    if not check_result:
        log.error(
            f"list_contain_check: List({str(superlist)}) does not contain list({str(sublist)})"
        )

    return check_result


def get_connect_object_name(_list):
    """ get the name of the objects that returned by the connection """
    if not isinstance(_list, list):
        log.error(
            f"[GET_CONNECT_OBJECT_NAME] Type of list({str(_list)}) is not a list."
        )
        return _list

    new_list = []
    for i in _list:
        if not isinstance(i, tuple):
            log.error(
                f"[GET_CONNECT_OBJECT_NAME] The element:{str(i)} of the list is not tuple, please check manually."
            )
            return _list

        if len(i) != 2:
            log.error(
                f"[GET_CONNECT_OBJECT_NAME] The length of the tuple:{str(i)} is not equal to 2, please check manually."
            )
            return _list

        if i[1] is not None:
            _obj_name = type(i[1]).__name__
            new_list.append((i[0], _obj_name))
        else:
            new_list.append(i)

    log.debug(
        f"[GET_CONNECT_OBJECT_NAME] list:{str(_list)} is reset to list:{new_list}"
    )
    return new_list


def equal_entity(exp, actual):
    """
    compare two entities containing vector field
    {"int64": 0, "float": 0.0, "float_vec": [0.09111554112502457, ..., 0.08652634258062468]}
    :param exp: exp entity
    :param actual: actual entity
    :return: bool
    """
    assert actual.keys() == exp.keys()
    for field, value in exp.items():
        if isinstance(value, list):
            assert len(actual[field]) == len(exp[field])
            for i in range(0, len(exp[field]), 4):
                assert abs(actual[field][i] - exp[field][i]) < ct.epsilon
        else:
            assert actual[field] == exp[field]
    return True


def entity_in(entity, entities, primary_field):
    """
    according to the primary key to judge entity in the entities list
    :param entity: dict
            {"int": 0, "vec": [0.999999, 0.111111]}
    :param entities: list of dict
            [{"int": 0, "vec": [0.999999, 0.111111]}, {"int": 1, "vec": [0.888888, 0.222222]}]
    :param primary_field: collection primary field
    :return: True or False
    """
    primary_default = ct.default_int64_field_name
    primary_field = primary_default if primary_field is None else primary_field
    primary_key = entity.get(primary_field, None)
    primary_keys = [e[primary_field] for e in entities]
    if primary_key not in primary_keys:
        return False
    index = primary_keys.index(primary_key)
    return equal_entity(entities[index], entity)


def remove_entity(entity, entities, primary_field):
    """
    according to the primary key to remove an entity from an entities list
    :param entity: dict
            {"int": 0, "vec": [0.999999, 0.111111]}
    :param entities: list of dict
            [{"int": 0, "vec": [0.999999, 0.111111]}, {"int": 1, "vec": [0.888888, 0.222222]}]
    :param primary_field: collection primary field
    :return: entities of removed entity
    """
    primary_default = ct.default_int64_field_name
    primary_field = primary_default if primary_field is None else primary_field
    primary_key = entity.get(primary_field, None)
    primary_keys = [e[primary_field] for e in entities]
    index = primary_keys.index(primary_key)
    entities.pop(index)
    return entities


def equal_entities_list(exp, actual, primary_field, with_vec=False):
    """
    compare two entities lists in inconsistent order
    :param with_vec: whether entities with vec field
    :param exp: exp entities list, list of dict
    :param actual: actual entities list, list of dict
    :return: True or False
    example:
    exp = [{"int": 0, "vec": [0.999999, 0.111111]}, {"int": 1, "vec": [0.888888, 0.222222]}]
    actual = [{"int": 1, "vec": [0.888888, 0.222222]}, {"int": 0, "vec": [0.999999, 0.111111]}]
    exp = actual
    """
    exp = exp.copy()
    if len(exp) != len(actual):
        return False

    if with_vec:
        for a in actual:
            # if vec field returned in query res
            if entity_in(a, exp, primary_field):
                try:
                    # if vec field returned in query res
                    remove_entity(a, exp, primary_field)
                except Exception as ex:
                    log.error(ex)
    else:
        for a in actual:
            if a in exp:
                try:
                    exp.remove(a)
                except Exception as ex:
                    log.error(ex)
    return len(exp) == 0


def output_field_value_check(search_res, original):
    """
    check if the value of output fields is correct
    :param search_res: the search result of specific output fields
    :param original: the data in the collection
    :return: True or False
    """
    limit = len(search_res[0])
    for i in range(limit):
        entity = eval(str(search_res[0][i]).split('entity: ', 1)[1])
        _id = search_res[0][i].id
        for field in entity.keys():
            if isinstance(entity[field], list):
                for order in range(0, len(entity[field]), 4):
                    assert abs(original[field][_id][order] - entity[field][order]) < ct.epsilon
            else:
                num = original[original[ct.default_int64_field_name] == _id].index.to_list()[0]
                assert original[field][num] == entity[field]

    return True
